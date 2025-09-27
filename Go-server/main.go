package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"runtime"
	"strings"
	"time"

	"github.com/coder/websocket"
)

type SystemInfo struct{
	Username string `json:"username"`
	Hostname string `json:"hostname"`
}

func getSystemInfo() SystemInfo{
	info := SystemInfo{
		Username: "user",
		Hostname: "web-cli",
	}

	if currentUser, err := user.Current(); err == nil{
		info.Username = currentUser.Username
	}

	if hostname, err := os.Hostname(); err == nil {
		info.Hostname = hostname
	}
	log.Printf("user: %s",info.Username);
	return info
}

func executeCommand(command string) string {
	command = strings.TrimSpace(command)
	if command == "" {
		return ""
	}

	log.Printf("Executing command: %s", command)
	
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return "Error: empty command"
	}
	
	cmdName := parts[0]
	args := parts[1:]
	
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()
	
	log.Printf("OS: %s", runtime.GOOS)
	
	cmd := exec.CommandContext(ctx, cmdName, args...)

	output, err := cmd.CombinedOutput()

	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Sprintf("Error: command '%s' timed out after 60 seconds", command)
		}

		if strings.Contains(err.Error(), "not found") ||
			strings.Contains(err.Error(), "cannot find") ||
			strings.Contains(err.Error(), "executable file not found") {
			return fmt.Sprintf("Error: command '%s' not found", cmdName)
		}

		return fmt.Sprintf("Error executing '%s': %s", command, err.Error())
	}

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

	sysInfo := getSystemInfo()
    infoJSON, _ := json.Marshal(sysInfo)
    err = conn.Write(context.Background(), websocket.MessageText, []byte("__SYSTEM_INFO__:"+string(infoJSON)))
    if err != nil {
        log.Printf("Failed to send system info: %v", err)
        return
    }

	ctx, cancel := context.WithTimeout(context.Background(), time.Hour*24)
	defer cancel()

	for {
		_, message, err := conn.Read(ctx)
		if err != nil {
			log.Printf("Read error: %v", err)
			break
		}

		input := strings.TrimSpace(string(message))
		
        if input == "\x03" {
            continue
        }

        if input != "" {
            output := executeCommand(input)

            // Send command output back to client
            err = conn.Write(ctx, websocket.MessageText, []byte(output))
            if err != nil {
                log.Printf("Write error: %v", err)
                break
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
