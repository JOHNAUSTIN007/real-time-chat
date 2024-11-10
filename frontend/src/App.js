import React, { useState, useEffect, useRef } from 'react';
import './App.css';

function App() {
  const [username, setUsername] = useState('');
  const [connected, setConnected] = useState(false);
  const [message, setMessage] = useState('');
  const [chat, setChat] = useState([]);
  const ws = useRef(null);

  useEffect(() => {
    return () => {
      if (ws.current) ws.current.close();
    };
  }, []);

  const connectWebSocket = () => {
    ws.current = new WebSocket('ws://localhost:8080/ws');

    ws.current.onopen = () => {
      console.log('WebSocket connected');
      setConnected(true);
    };

    ws.current.onmessage = (evt) => {
      const receivedMsg = JSON.parse(evt.data);
      setChat((prev) => [...prev, receivedMsg]);
    };

    ws.current.onclose = () => {
      console.log('WebSocket disconnected');
      setConnected(false);
    };

    ws.current.onerror = (err) => {
      console.error('WebSocket error:', err);
    };
  };

  const sendMessage = () => {
    if (ws.current && connected && message.trim() !== '') {
      const msg = {
        username: username || 'Anonymous',
        message: message,
      };
      ws.current.send(JSON.stringify(msg));
      setMessage('');
    }
  };

  const handleKeyPress = (e) => {
    if (e.key === 'Enter') {
      sendMessage();
    }
  };

  return (
    <div className="App">
      {!connected ? (
        <div className="login">
          <h2>Enter your username</h2>
          <input
            type="text"
            value={username}
            onChange={(e) => setUsername(e.target.value)}
            onKeyPress={(e) => {
              if (e.key === 'Enter') connectWebSocket();
            }}
            placeholder="Username"
          />
          <button onClick={connectWebSocket}>Join Chat</button>
        </div>
      ) : (
        <div className="chat-container">
          <div className="chat-header">
            <h2>Real-Time Chat</h2>
            <button onClick={() => ws.current.close()}>Disconnect</button>
          </div>
          <div className="chat-messages">
            {chat.map((msg, index) => (
              <div key={index} className="message">
                <strong>{msg.username}: </strong>
                <span>{msg.message}</span>
              </div>
            ))}
          </div>
          <div className="chat-input">
            <input
              type="text"
              value={message}
              onChange={(e) => setMessage(e.target.value)}
              onKeyPress={handleKeyPress}
              placeholder="Type your message..."
            />
            <button onClick={sendMessage}>Send</button>
          </div>
        </div>
      )}
    </div>
  );
}

export default App;
