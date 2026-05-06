# Web CLI

Browser-based interactive terminal powered by Go, PTY, and xterm.js.

## Architecture

```
Browser (xterm.js) ──WebSocket──▶ Go Server ──PTY──▶ /bin/bash
                                   │
                                   ├── HTTP handler (upgrade)
                                   ├── Service layer (lifecycle)
                                   └── Domain layer (PTY wrapper)
```

Single binary serves both the frontend and the WebSocket backend. No separate frontend build step.

## Project Structure

```
Go-server/
├── main.go                                        # App entry: config → services → router → server
├── go.mod / go.sum
├── internal/
│   ├── config/
│   │   └── config.go                              # Environment variable management
│   ├── domain/
│   │   └── shell.go                               # PTY wrapper (NewShell, Read, Write, Resize, Close)
│   ├── usecase/
│   │   └── shell_service.go                       # PTY lifecycle orchestration with thread safety
│   └── infrastructure/
│       └── http/
│           ├── handler_shell_ws.go                 # WebSocket ↔ PTY bridge (bi-directional)
│           ├── handler_template.go                 # HTML template rendering
│           └── router.go                           # Routes + CORS/logging middleware
└── web/
    ├── index.html                                  # Terminal UI (CDN xterm.js)
    └── js/
        └── main.js                                 # WebSocket client with PTY protocol
```

## Quick Start

```bash
cd Go-server
go run main.go
```

Open `http://localhost:8080` in your browser. You'll get an interactive bash shell.

## Configuration

Environment variables with defaults:

| Variable | Default | Description |
|----------|---------|-------------|
| `HOST` | `0.0.0.0` | Server bind address |
| `PORT` | `8080` | Server port |

```bash
PORT=9090 go run main.go
```

## WebSocket Protocol

JSON messages over WebSocket:

**Client → Server:**
```json
{"type": "input", "data": "ls -la\n"}
{"type": "resize", "rows": 40, "cols": 120}
```

**Server → Client:**
```json
{"type": "output", "data": "terminal output..."}
{"type": "error", "data": "error message"}
```

## Dependencies

| Package | Purpose |
|---------|---------|
| `github.com/coder/websocket` | WebSocket server (context-aware) |
| `github.com/creack/pty` | PTY allocation for interactive shell |
| `xterm.js` (CDN) | Terminal emulator in browser |

## Requirements

- Go 1.24+
- Linux (PTY support required)
- `/bin/bash` available

## Graceful Shutdown

`Ctrl+C` triggers graceful shutdown — active PTY sessions are cleaned up within 10 seconds with no orphaned processes.

## Limitations

- **Single session** — only one WebSocket client can have an active shell at a time
- **No authentication** — anyone with access can get a shell (local/trusted use only)
- **No persistence** — container filesystem is ephemeral on cloud deployments
