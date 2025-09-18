# WebSocket Server

This project implements a simple WebSocket server in Go for real-time communication.

## Features

- Handles multiple WebSocket connections.
- Broadcasts messages to all connected clients.
- Simple and easy to understand code structure.

## Requirements

- Go 1.16 or later

## Installation

1. Clone the repository:
   ```
   git clone <repository-url>
   cd websocket-server
   ```

2. Install dependencies:
   ```
   go mod tidy
   ```

## Running the Server

To run the WebSocket server, execute the following command:

```
go run main.go
```

The server will start listening on `localhost:8080`.

## Usage

Connect to the WebSocket server using a WebSocket client. You can use browser-based tools like the WebSocket King Client or write your own client in JavaScript.

## License

This project is licensed under the MIT License.