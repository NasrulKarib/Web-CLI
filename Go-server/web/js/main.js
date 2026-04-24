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
let username = 'host';
let hostname = 'ubuntu';

let commandHistory = [];
let historyIndex = -1;
let isCommandRunning = false;

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

function handleOutputMessage(data){
  const msg = JSON.parse(data)

  try{
    
    switch (msg.type) {
          case 'stdout':
              if (msg.content.trim() !== '') {
                  term.write('\r\n' + msg.content);
              }
              break;
                
          case 'stderr':
              if (msg.content.trim() !== '') {
                  term.write('\r\n\x1b[31m' + msg.content + '\x1b[0m');
              }
              break;
                
          case 'status':
              if (msg.content.trim() !== '') {
                  term.write('\r\n\x1b[33m[' + msg.content + ']\x1b[0m');
              }
              break;
                
          case 'system':
              if (msg.content === '__COMMAND_COMPLETE__') {
                  isCommandRunning = false;
                  term.write('\r\n');
                  writePrompt();
                  currentLine = '';
                  historyIndex = -1; 
                  return;
              } else {
                  term.write('\r\n\x1b[36m' + msg.content + '\x1b[0m');
              }
              break;
                
          default:
              term.write('\r\n' + msg.content);
              break;
        }
  } catch (e) {
    if (data.trim() !== '') {
        term.write(data);
    }
  }
}

function connectWebSocket(){
  try {
    ws = new WebSocket('ws://localhost:8080/ws')

    ws.onopen = async(event) =>{
      await typewriterEffect('✓ Connected to WebSocket server\n', '38;5;120', 30);
      term.writeln('');
      writePrompt();
      currentLine = '';
      isCommandRunning = false;
    }

    ws.onmessage = (event)=>{
      const data = event.data
      if(data.startsWith('__SYSTEM_INFO__:')){
        const info = JSON.parse(data.replace('__SYSTEM_INFO__:',''));
        username = info.username || 'user';
        hostname = info.hostname || 'web-cli';
        return;
      }

      handleOutputMessage(data)
    }

    ws.onclose = (event) =>{
      isCommandRunning = false;
      term.writeln('\r\n\x1b[31m✗ Disconnected from server\x1b[0m');
      term.writeln(`Close code: ${event.code}, Reason: ${event.reason || 'Unknown'}`);
      
      setTimeout(() => {
          term.writeln('Attempting to reconnect...');
          connectWebSocket();
      }, 3000);
    }

    ws.onerror = (error) => {
      isCommandRunning = false;
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
  
  if(isCommandRunning && charCode !== 3) {
    // prevent inp during command exectution
    return;
  }

  if (charCode === 13) {  
    if(ws && ws.readyState === WebSocket.OPEN && currentLine.trim() !== '') {
      commandHistory.push(currentLine.trim())

      // limit to latest 50 commands
      if(commandHistory.length > 50){
        commandHistory.shift();
      }

      isCommandRunning = true;
      ws.send(currentLine);
      currentLine = '';
      historyIndex = -1;
    } else if (currentLine.trim() === '') {
      term.write('\r\n');
      writePrompt();
    }
  } 
  
  else if (charCode == 8 || charCode == 127) {
    if(currentLine.length > 0){
      currentLine = currentLine.slice(0, -1);
      term.write('\b \b');
    }
  } 

  else if (charCode === 27) {
    const seq = e.slice(1);
    if (seq === '[A') { // Up arrow
        navigateHistory(-1);
    } else if (seq === '[B') { // Down arrow
        navigateHistory(1);
    }
  }
  
  else if(charCode >= 32 && charCode <= 126){
    currentLine += e
    term.write(e);
  }

  else if (charCode === 3) { 
    term.write('^C');
    if(isCommandRunning) {
      isCommandRunning = false;
      
      if (ws && ws.readyState === WebSocket.OPEN) {
          ws.send('\x03'); 
      }
    } 

    currentLine = '';
    term.write('\r\n');
    writePrompt();
  }
});

function navigateHistory(direction) {
  if(commandHistory.length === 0) return;

  term.write('\r\x1b[K');
  writePrompt();

  if(direction === -1) {
    // up
    if(historyIndex === -1) {
      historyIndex = commandHistory.length - 1;
    } else if (historyIndex > 0){
      historyIndex--;
    }
  } else if (direction === 1) {
    // down
    if (historyIndex === -1) {
      return;
    } else if (historyIndex < commandHistory.length - 1) {
      historyIndex++;
    } else {
      historyIndex = -1;
      currentLine = '';
      return;
    }
  }

  if(historyIndex >= 0 && historyIndex < commandHistory.length){
    currentLine = commandHistory[historyIndex]
    term.write(currentLine)
  }
}

window.addEventListener('resize', () => {
    fitAddon.fit();
});

term.writeln('\x1b[36m╔══════════════════════════════════════╗\x1b[0m');
term.writeln('\x1b[36m║        Web-based CLI Terminal        ║\x1b[0m');
term.writeln('\x1b[36m╚══════════════════════════════════════╝\x1b[0m');
term.writeln('');
term.writeln('\x1b[33mFeatures:\x1b[0m');
term.writeln('• \x1b[32m↑/↓ arrows\x1b[0m - Navigate command history (50 commands)');
term.writeln('• \x1b[31mRed text\x1b[0m - Error messages and stderr');
term.writeln('• \x1b[37mWhite text\x1b[0m - Normal command output');
term.writeln('• \x1b[33mYellow text\x1b[0m - Status and loading messages');
term.writeln('• \x1b[36mCyan text\x1b[0m - System notifications');
term.writeln('• \x1b[32mCtrl+C\x1b[0m - Interrupt running commands');
term.writeln('');
term.writeln('Connecting to Go WebSocket server...\n');

connectWebSocket();
