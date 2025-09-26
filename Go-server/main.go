package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/coder/websocket"
)

func executeCommand(command string) string {
	command = strings.TrimSpace(command)
	if command == "" {
		return ""
	}

	log.Printf("Executing command: %s", command)
	
	// Parse command and arguments
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return "Error: empty command"
	}
	
	cmdName := parts[0]
	args := parts[1:]
	
	// Create command with 10-second timeout for safety
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	log.Printf("OS: %s", runtime.GOOS)
	// Create the command
	cmd := exec.CommandContext(ctx, cmdName, args...)

	// Execute and capture both stdout and stderr
	output, err := cmd.CombinedOutput()

	// Handle different error cases
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Sprintf("Error: command '%s' timed out after 10 seconds", command)
		}

		// Check for "command not found" errors
		if strings.Contains(err.Error(), "not found") ||
			strings.Contains(err.Error(), "cannot find") ||
			strings.Contains(err.Error(), "executable file not found") {
			return fmt.Sprintf("Error: command '%s' not found", cmdName)
		}

		// Return the actual error for other cases
		return fmt.Sprintf("Error executing '%s': %s", command, err.Error())
	}

	// Return command output
	if len(output) == 0 {
		return fmt.Sprintf("Command '%s' executed successfully (no output)", command)
	}

	return string(output)
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		InsecureSkipVerify: true,
	})
	if err != nil {
		log.Printf("WebSocket accept failed: %v", err)
		return
	}
	defer conn.Close(websocket.StatusInternalError, "Server error")

	log.Println("Client connected")

	ctx, cancel := context.WithTimeout(context.Background(), time.Hour*24)
	defer cancel()

	var commandBuffer strings.Builder
	for {
		// Read message from client
		_, message, err := conn.Read(ctx)
		if err != nil {
			log.Printf("Read error: %v", err)
			break
		}

		input := string(message)

		switch input {
		case "\n": // Enter key pressed - execute command
			command := strings.TrimSpace(commandBuffer.String())
			if command != "" {
				// Execute the accumulated command
				output := executeCommand(command)

				// Send command output back to client
				err = conn.Write(ctx, websocket.MessageText, []byte(output))
				if err != nil {
					log.Printf("Write error: %v", err)
					break
				}
			}
			// Clear the command buffer after execution
			commandBuffer.Reset()

		case "\b": // Backspace key
			// Remove last character from buffer
			str := commandBuffer.String()
			if len(str) > 0 {
				commandBuffer.Reset()
				commandBuffer.WriteString(str[:len(str)-1])
			}

		default:
			// Regular character - add to command buffer
			if len(input) == 1 && input[0] >= 32 && input[0] <= 126 {
				commandBuffer.WriteString(input)
			}
		}
	}

	log.Println("Client disconnected")
	conn.Close(websocket.StatusNormalClosure, "Connection closed")
}

func enableCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

func homePage(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	w.Write([]byte("Websocket server is running!"))
}

func main() {
	// Route
	http.HandleFunc("/", homePage)
	http.HandleFunc("/ws", handleWebSocket)

	fmt.Println("WebSocket server starting on :8080")
	fmt.Println("WebSocket endpoint: ws://localhost:8080/ws")

	log.Fatal(http.ListenAndServe(":8080", nil))
}
