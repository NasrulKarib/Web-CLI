package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/coder/websocket"
)

type SystemInfo struct{
	Username string `json:"username"`
	Hostname string `json:"hostname"`
}

type OutputMessage struct{
	Type string `json:"type"`
	Content string `json:"content"`
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
	return info
}

func executeCommand(command string, conn *websocket.Conn) {
    log.Printf("Executing command: %s", command)
    
    if strings.TrimSpace(command) == "" {
        sendMessage(conn, "system", "Error: empty command")
        return
    }
    
    cmdCtx, cmdCancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cmdCancel()
        
    var cmd *exec.Cmd
    if runtime.GOOS == "windows" {
        cmd = exec.CommandContext(cmdCtx, "cmd", "/C", command)
    } else {
        cmd = exec.CommandContext(cmdCtx, "/bin/sh", "-c", command)
    }

    stdoutPipe, err := cmd.StdoutPipe()
    if err != nil {
        sendMessage(conn, "stderr", fmt.Sprintf("Error creating stdout pipe: %v", err))
        return
    }

    stderrPipe, err := cmd.StderrPipe()
    if err != nil{
        sendMessage(conn, "stderr", fmt.Sprintf("Error creating stderr pipe: %v", err))
        return
    }
    if err := cmd.Start(); err != nil {
        sendMessage(conn, "stderr", fmt.Sprintf("Error starting command: %v", err))
        return
    }

    var wg sync.WaitGroup
    wg.Add(2)

    go func(){
        defer wg.Done()
        streamOutput(stdoutPipe, conn, "stdout")
    }()

    go func(){
        defer wg.Done()
        streamOutput(stderrPipe, conn, "stderr")
    }()
    
    
    wg.Wait()
        
	err = cmd.Wait()

        
    if err != nil {
        if cmdCtx.Err() == context.DeadlineExceeded {
            sendMessage(conn, "stderr", fmt.Sprintf("Command '%s' timed out after 30 seconds", command))
        } 
    }
    
    sendMessage(conn, "system", "__COMMAND_COMPLETE__")
}

func streamOutput(pipe io.ReadCloser, conn *websocket.Conn, outputType string) {
    reader := bufio.NewReader(pipe)
    buf := make([]byte, 1024) // chunk size = 1024 byte
    for {
        n, err := reader.Read(buf)
        if n > 0 {
            sendMessage(conn, outputType, string(buf[:n]))
        }
        if err != nil {
            if err != io.EOF {
                log.Printf("Error reading %s: %v", outputType, err)
            }
            break
        }
    }
}


func sendMessage(conn *websocket.Conn, msgType, content string) {

    msg := OutputMessage{
        Type: msgType,
        Content: content,
    }

    msgJSON, err := json.Marshal(msg)
    if err != nil {
        log.Printf("Error marshaling message: %v", err)
        return
    }

    err = conn.Write(context.Background(), websocket.MessageText, msgJSON)
    if err != nil {
        log.Printf("Error sending message: %v", err)
    }
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

	sessionCtx, sessionCancel := context.WithTimeout(context.Background(), time.Hour*24)
	defer sessionCancel()

	for {
		_, message, err := conn.Read(sessionCtx)
		if err != nil {
			log.Printf("Read error: %v", err)
			sessionCancel()
			break
		}

		input := strings.TrimSpace(string(message))
		
        if input == "\x03" {
			log.Println("Received Ctrl+C from client")
            continue
        }

        if input != "" {
            go executeCommand(input, conn)
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
