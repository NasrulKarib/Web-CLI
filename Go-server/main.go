package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/coder/websocket"
)

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

	for {
		// Read message from client
		_, message, err := conn.Read(ctx)
		if err != nil {
			log.Printf("Read error: %v", err)
			break
		}

		fmt.Printf("Client said: %s\n", string(message))

		response := fmt.Sprintf("Server: %s", string(message))

		// Write response back to client
		err = conn.Write(ctx, websocket.MessageText, []byte(response))
		if err != nil {
			log.Printf("Write error: %v", err)
			break
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
