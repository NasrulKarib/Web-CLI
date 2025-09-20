# WebSocket Real-Time Communication Project

This project demonstrates real-time communication between a web client and Go server using WebSocket technology.

## Project Overview

The project consists of two main components:
- **Go Server**: WebSocket server that handles connections and echoes messages
- **Web Client**: HTML/CSS/JavaScript frontend for user interaction

## Features

- ✅ Real-time bidirectional communication via WebSocket
- ✅ Client sends messages to Go server
- ✅ Server echoes messages back to client
- ✅ Connection status indicators
- ✅ Clean separation between frontend and backend
- ✅ CORS enabled for cross-origin requests
- ✅ Interactive web interface with connect/disconnect functionality

## Project Structure

```
Websocket with Go/
├── Go server/
│   ├── main.go          # WebSocket server implementation
│   ├── go.mod           # Go module dependencies
│   └── go.sum           # Dependency checksums
├── Client/
│   ├── index.html       # Web interface
│   ├── client.js        # WebSocket client logic
│   └── style.css        # Styling
└── README.md
```

## Requirements

- Go 1.16 or later
- Modern web browser with WebSocket support

## Installation & Setup

1. **Navigate to the project directory:**
   ```bash
   cd "c:\Programming\Websocket with Go"
   ```

2. **Install Go dependencies:**
   ```bash
   cd "Go server"
   go mod tidy
   ```

## Running the Application

### 1. Start the Go Server
```bash
cd "Go server"
go run main.go
```
Server will start on `http://localhost:8080`
WebSocket endpoint: `ws://localhost:8080/ws`

### 2. Open the Client
Navigate to the `Client` folder and open `index.html` in your web browser, or serve it using:
```bash
cd Client
python -m http.server 3000
```
Then open `http://localhost:3000`

## How It Works

1. **Client Connection**: Web client connects to WebSocket server at `ws://localhost:8080/ws`
2. **Message Flow**: 
   - User types message in input field
   - Client sends message via WebSocket
   - Server receives message and prints "Client said: [message]"
   - Server echoes back "Echo: [message]"
   - Client displays the echo response
3. **Real-time Communication**: Messages are exchanged instantly without page refresh

## Usage Example

1. Open the web client in your browser
2. Click "Connect" to establish WebSocket connection
3. Type "hello" in the input field
4. Click "Send" or press Enter
5. See the server response: "Echo: hello"

## Dependencies

- **Go Server**: `github.com/gorilla/websocket` - WebSocket implementation
- **Client**: Pure HTML/CSS/JavaScript - No external dependencies

## License

This project is licensed under the MIT License.