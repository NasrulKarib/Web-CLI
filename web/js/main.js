(function () {
    'use strict';

    var term = new Terminal({
        cursorBlink: true,
        fontSize: 14,
        fontFamily: '"Cascadia Code", "Fira Code", monospace',
        theme: {
            background: '#0d0d0d',
            foreground: '#f0f0f0',
            cursor:     '#f0f0f0',
        },
    });

    var fitAddon = new FitAddon.FitAddon();
    term.loadAddon(fitAddon);
    term.open(document.getElementById('terminal-container'));
    fitAddon.fit();

    var statusEl = document.getElementById('status');
    var ws = null;
    var reconnectTimer = null;

    function connect() {
        clearTimeout(reconnectTimer);

        var proto = location.protocol === 'https:' ? 'wss' : 'ws';
        var wsUrl = proto + '://' + location.host + '/ws?rows=' + term.rows + '&cols=' + term.cols;

        ws = new WebSocket(wsUrl);

        ws.onopen = function () {
            statusEl.textContent = 'Connected';
            statusEl.className = 'connected';
        };

        ws.onmessage = function (event) {
            try {
                var msg = JSON.parse(event.data);
                if (msg.type === 'output') {
                    term.write(msg.data);
                } else if (msg.type === 'error') {
                    term.write('\r\n\x1b[31mError: ' + msg.data + '\x1b[0m\r\n');
                }
            } catch (e) {
                console.error('Failed to parse message:', e);
            }
        };

        ws.onerror = function (err) {
            console.error('WebSocket error:', err);
        };

        ws.onclose = function () {
            statusEl.textContent = 'Disconnected';
            statusEl.className = 'disconnected';
            reconnectTimer = setTimeout(connect, 3000);
        };
    }

    term.onData(function (data) {
        if (ws && ws.readyState === WebSocket.OPEN) {
            ws.send(JSON.stringify({ type: 'input', data: data }));
        }
    });

    var resizeObserver = new ResizeObserver(function () {
        fitAddon.fit();
        if (ws && ws.readyState === WebSocket.OPEN) {
            ws.send(JSON.stringify({
                type: 'resize',
                rows: term.rows,
                cols: term.cols,
            }));
        }
    });
    resizeObserver.observe(document.getElementById('terminal-container'));

    connect();
}());
