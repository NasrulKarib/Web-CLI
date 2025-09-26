# Web-Based CLI Project – Copilot Instruction

## Project Overview

You are assisting in building a **web-based terminal/CLI** using **Go** for the backend and **xterm.js** for the frontend.
The project goal is to allow users to interact with a **remote shell via browser**, with **real-time bi-directional communication** using WebSockets.
Security, concurrency, and proper session management should be considered.

---

## Tech Stack

* **Backend:** Go

  * WebSocket library: `github.com/coder/websocket`
  * Process handling: `os/exec` for commands, `creack/pty` for interactive shells
  * HTTP server: `net/http`
* **Frontend:**

  * `xterm.js` for terminal emulator
  * Vanilla JS or minimal HTML/CSS
  * WebSocket API for client-server communication
* **Optional Tools:** Docker for sandboxing, JWT/session for auth, logging

---

## Current State

* WebSocket connection works:

  * Server upgrades HTTP to WebSocket (`/ws`)
  * Echo functionality implemented (server receives and echoes messages)
  * Proper error handling, graceful close, CORS support
* Client side:

  * Connects to server, sends and receives messages
  * Real-time status updates (Connected/Disconnected)
  * Enter key support to send messages
  * Auto-connect on page load, message history, clean UI

---

## Project Goals 

1. **Integrate xterm.js**:

   * Replace input box with terminal emulator.
   * Send each keystroke from xterm.js to server via WebSocket.
   * Display server responses back in xterm.js.

2. **Interactive Shell on Server**:

   * Use `creack/pty` to spawn `/bin/bash` or `/bin/sh`.
   * Connect PTY stdin/stdout to WebSocket.
   * Support interactive commands (like `vim`, `top`, `nano`).

3. **Multi-User Support**:

   * Each WebSocket connection gets its own PTY session.
   * Assign session IDs and track active sessions.

4. **Security / Sandbox**:

   * Optional: run PTY inside Docker or chroot for isolation.
   * Implement authentication (JWT/session) to restrict access.
   * Limit resource usage per session (memory, CPU).

5. **Enhancements**:

   * Terminal resizing support
   * Copy/paste functionality
   * Logging of commands and outputs for audit

6. Custom commands (ls, grep, pwd, etc.)
7. Real-time terminal updates
8. User-friendly error messages
9. Optional autocomplete for commands

---

## Copilot Instructions / Usage

When writing code:

* Follow Go best practices (error handling, concurrency safety).
* Avoid using external JS frameworks other than xterm.js.
* Maintain **real-time streaming**, don’t buffer full output.
* Code should be modular:

  * `main.go` → entry point
  * `handlers/` → WebSocket handlers
  * `terminal/` → PTY/session management
* Provide comments explaining **what each block does**.
* Keep security in mind (don’t execute unsanitized input directly).

---

## Example Flow

1. Browser opens terminal → WebSocket connection established.
2. User types `ls -la` in xterm.js → sends to Go server.
3. Server executes command in PTY → streams output line by line.
4. Browser receives output → writes to xterm.js.
5. User disconnects → server cleans up PTY session.
