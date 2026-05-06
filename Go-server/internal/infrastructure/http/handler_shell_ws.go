package http

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"sync"

	"github.com/coder/websocket"

	"websocket-server/internal/usecase"
)

// inboundMsg represents a message from the client.
type inboundMsg struct {
	Type string `json:"type"`
	Data string `json:"data"`
	Rows uint16 `json:"rows,omitempty"`
	Cols uint16 `json:"cols,omitempty"`
}

// outboundMsg represents a message sent to the client.
type outboundMsg struct {
	Type string `json:"type"`
	Data string `json:"data"`
}

// ShellWSHandler handles WebSocket connections for PTY shell sessions.
type ShellWSHandler struct {
	shellService *usecase.ShellService
}

// NewShellWSHandler creates a new ShellWSHandler with the given ShellService.
func NewShellWSHandler(shellService *usecase.ShellService) *ShellWSHandler {
	return &ShellWSHandler{shellService: shellService}
}

// Handle upgrades an HTTP connection to WebSocket and bridges it with a PTY shell.
func (h *ShellWSHandler) Handle(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		OriginPatterns: []string{"*"},
	})
	if err != nil {
		log.Printf("websocket: upgrade failed: %v", err)
		return
	}
	defer conn.Close(websocket.StatusNormalClosure, "session ended")

	rows, cols := parseTerminalSize(r.URL.Query())

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	if err := h.shellService.Start(ctx, rows, cols); err != nil {
		sendError(conn, ctx, err.Error())
		log.Printf("websocket: failed to start shell: %v", err)
		return
	}
	defer func() {
		if err := h.shellService.Stop(); err != nil {
			log.Printf("websocket: shell stop error: %v", err)
		}
	}()

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		defer cancel()
		h.pumpInput(ctx, conn)
	}()

	go func() {
		defer wg.Done()
		defer cancel()
		h.pumpOutput(ctx, conn)
	}()

	wg.Wait()
}

// pumpInput reads messages from the WebSocket and forwards them to the PTY.
func (h *ShellWSHandler) pumpInput(ctx context.Context, conn *websocket.Conn) {
	for {
		_, data, err := conn.Read(ctx)
		if err != nil {
			if ctx.Err() == nil {
				log.Printf("websocket: input read error: %v", err)
			}
			return
		}

		var msg inboundMsg
		if err := json.Unmarshal(data, &msg); err != nil {
			log.Printf("websocket: invalid message (ignoring): %v", err)
			continue
		}

		switch msg.Type {
		case "input":
			if _, err := h.shellService.Write([]byte(msg.Data)); err != nil {
				log.Printf("websocket: PTY write error: %v", err)
				return
			}
		case "resize":
			if err := h.shellService.Resize(msg.Rows, msg.Cols); err != nil {
				log.Printf("websocket: PTY resize error: %v", err)
			}
		default:
			log.Printf("websocket: unknown message type %q (ignoring)", msg.Type)
		}
	}
}

// pumpOutput reads PTY output and forwards it to the WebSocket.
// Uses ReadContext so the pump exits immediately on context cancellation,
// even if the underlying PTY read is blocked.
func (h *ShellWSHandler) pumpOutput(ctx context.Context, conn *websocket.Conn) {
	for {
		output, err := h.shellService.ReadContext(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return
			}
			if errors.Is(err, io.EOF) {
				return
			}
			log.Printf("websocket: PTY read error: %v", err)
			return
		}

		msg, _ := json.Marshal(outboundMsg{
			Type: "output",
			Data: string(output),
		})

		if err := conn.Write(ctx, websocket.MessageText, msg); err != nil {
			if ctx.Err() == nil {
				log.Printf("websocket: output write error: %v", err)
			}
			return
		}
	}
}

// parseTerminalSize extracts rows and cols from query parameters with defaults.
func parseTerminalSize(query url.Values) (rows, cols uint16) {
	rows, cols = 24, 80

	if r := query.Get("rows"); r != "" {
		if val, err := strconv.Atoi(r); err == nil && val > 0 && val <= 500 {
			rows = uint16(val)
		}
	}

	if c := query.Get("cols"); c != "" {
		if val, err := strconv.Atoi(c); err == nil && val > 0 && val <= 500 {
			cols = uint16(val)
		}
	}

	return
}

// sendError sends a JSON error message to the WebSocket client.
func sendError(conn *websocket.Conn, ctx context.Context, message string) {
	msg, _ := json.Marshal(outboundMsg{Type: "error", Data: message})
	if err := conn.Write(ctx, websocket.MessageText, msg); err != nil {
		log.Printf("websocket: failed to send error message: %v", err)
	}
}
