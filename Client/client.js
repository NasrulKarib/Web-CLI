let ws = null;
const messagesDiv = document.getElementById('messages');
const statusDiv = document.getElementById('status');
const messageInput = document.getElementById('messageInput');

// WebSocket connection function
function connectWebSocket() {
    if (ws && ws.readyState === WebSocket.OPEN) {
        addMessage('Already connected!', 'system');
        return;
    }

    ws = new WebSocket('ws://localhost:8080/ws');
    
    ws.onopen = function() {
        updateStatus('Connected', true);
        addMessage('Connected to server', 'system');
    };
    
    ws.onmessage = function(event) {
        addMessage(event.data, 'received');
    };
    
    ws.onclose = function() {
        updateStatus('Disconnected', false);
        addMessage('Disconnected from server', 'system');
    };
    
    ws.onerror = function(error) {
        updateStatus('Connection Error', false);
        addMessage('Connection error: ' + error, 'error');
    };
}

// Disconnect WebSocket
function disconnectWebSocket() {
    if (ws) {
        ws.close();
        ws = null;
    }
}

// Send message function
function sendMessage() {
    const message = messageInput.value.trim();
    
    if (!message) {
        alert('Please enter a message!');
        return;
    }
    
    if (!ws || ws.readyState !== WebSocket.OPEN) {
        alert('WebSocket is not connected!');
        return;
    }
    
    // Send message to server
    ws.send(message);
    
    // Display sent message
    addMessage(`You: ${message}`, 'sent');
    
    // Clear input
    messageInput.value = '';
}

// Add message to display
function addMessage(message, type) {
    const messageElement = document.createElement('div');
    messageElement.className = `message ${type}`;
    messageElement.textContent = message;
    messagesDiv.appendChild(messageElement);
    messagesDiv.scrollTop = messagesDiv.scrollHeight;
}

// Update connection status
function updateStatus(status, connected) {
    statusDiv.textContent = status;
    statusDiv.className = connected ? 'status connected' : 'status disconnected';
}

// Enter key support
messageInput.addEventListener('keypress', function(e) {
    if (e.key === 'Enter') {
        sendMessage();
    }
});

// Auto-connect on page load
window.addEventListener('load', function() {
    connectWebSocket();
});