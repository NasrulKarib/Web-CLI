package main

import(
	"fmt"
	"log"
	"net/http"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func (r *http.Request) bool{
		return true;
	},
}

func handleWebSocket(w http.ResponseWriter, r *http.Request){
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil{
		log.Printf("WebSocket upgrade failed: %v", err)
        return
	}
	defer conn.Close()
	log.Println("Client connected")

	for{
		_, message, err := conn.ReadMessage()
		if err != nil{
			log.Printf("Read error: %v", err)
            break
		}

		fmt.Printf("Client said: %s\n",string(message))

		response := fmt.Sprintf("Server: %s", string(message))
		err = conn.WriteMessage(websocket.TextMessage, []byte(response))
		if err != nil{
			log.Printf("Write error: %v", err)
            break
		}
	}

	log.Println("Client disconnected")

}

func enableCORS(w http.ResponseWriter) {
    w.Header().Set("Access-Control-Allow-Origin", "*")
    w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
    w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

func homePage(w http.ResponseWriter, r *http.Request){
	enableCORS(w)
	w.Write([]byte("Websocket server is running!"))
}

func main(){
	// Route
	http.HandleFunc("/", homePage)
	http.HandleFunc("/ws",handleWebSocket)

	fmt.Println("WebSocket server starting on :8080")
    fmt.Println("WebSocket endpoint: ws://localhost:8080/ws")
    
    log.Fatal(http.ListenAndServe(":8080", nil))
}