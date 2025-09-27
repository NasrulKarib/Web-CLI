import { Terminal } from '@xterm/xterm';
import { FitAddon } from '@xterm/addon-fit';
import '@xterm/xterm/css/xterm.css';


const term = new Terminal({
    cols: 80,
    rows: 24,
    cursorBlink: true,
    allowProposedApi : true,
    theme: {
        background: '#000000',
        foreground: '#ffffff',
        cursor: '#ffffff',
        cursorAccent: '#000000',
        selectionBackground: '#ffffff',
        selectionForeground: '#000000'
    },
    fontFamily: 'Monaco, Menlo, "Ubuntu Mono", monospace',
    fontSize: 14,
    lineHeight: 1.2,
    convertEol: true,
});

const fitAddon = new FitAddon()
term.loadAddon(fitAddon);

term.open(document.getElementById("terminal"));
fitAddon.fit();

let ws = null
let currentLine = '';
let username = 'arian';
let hostname = 'ubuntu';


function writePrompt(){
  term.write(`\x1b[38;5;120m${username}@${hostname}:\x1b[0m$ `);
}

function typewriterEffect(text, color = '', speed = 50) {
    return new Promise((resolve) => {
        let i = 0;
        const colorCode = color ? `\x1b[${color}m` : '';
        const resetCode = color ? '\x1b[0m' : '';
        
        function typeChar() {
            if (i < text.length) {
                term.write(colorCode + text[i] + resetCode);
                i++;
                setTimeout(typeChar, speed);
            } else {
                resolve();
            }
        }
        typeChar();
    });
}

function connectWebSocket(){
  try {
    ws = new WebSocket('ws://localhost:8080/ws')

    ws.onopen = async(event) =>{
      await typewriterEffect('✓ Connected to WebSocket server\n', '38;5;120', 30);

      term.writeln('');
      writePrompt();
      currentLine = '';
    }

    ws.onmessage = (event)=>{
      const data = event.data

      if(data.startsWith('__SYSTEM_INFO__:')){
        const info = JSON.parse(data.replace('__SYSTEM_INFO__:',''));
        username = info.username || 'user';
        hostname = info.hostname || 'web-cli';
        return;
      }

      const output = data;
      if(output.trim() != ''){
        term.writeln('\r\n' + output);
      }
      writePrompt();
      currentLine = '';
    }

    ws.onclose = (event) =>{
      term.writeln('\r\n\x1b[31m✗ Disconnected from server\x1b[0m');
      term.writeln(`Close code: ${event.code}, Reason: ${event.reason || 'Unknown'}`);
      
      setTimeout(() => {
          term.writeln('Attempting to reconnect...');
          connectWebSocket();
      }, 3000);
    }

    ws.onerror = (error) => {
      term.writeln('\r\n\x1b[31m✗ WebSocket connection error\x1b[0m');
      console.error('WebSocket error:', error);
    };

  } catch (error){
    term.writeln('\x1b[31m✗ Failed to connect to WebSocket server\x1b[0m');
    console.error('Connection error:', error);
  }
}

term.onData(e => {
  const charCode = e.charCodeAt(0);
  
  if (charCode === 13) {  
    if(ws && ws.readyState === WebSocket.OPEN){
      ws.send(currentLine);
      currentLine = '';
    }
  } 
  
  else if (charCode == 8 || charCode == 127) {
    if(currentLine.length > 0){
      currentLine = currentLine.slice(0, -1);

      term.write('\b \b');
      
    }
  } 
  
  else if(charCode >= 32 && charCode <= 126){
    currentLine += e

    term.write(e);
  }

  else if (charCode === 3) { 
    term.write('^C');
    currentLine = '';
    if (ws && ws.readyState === WebSocket.OPEN) {
        ws.send('\x03'); 
    }
  }
});

window.addEventListener('resize', () => {
    fitAddon.fit();
});


term.writeln('Web-based CLI Terminal');
term.writeln('Connecting to Go WebSocket server...\n');
connectWebSocket();
