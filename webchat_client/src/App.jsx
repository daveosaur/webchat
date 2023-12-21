import { useState, useRef, useEffect } from 'react'
import './App.css'
import ChatBox from './chatbox.jsx'

const starterNames = ["guy", "bobby", "dave", "notdave", "gooble", "borb"];

function App() {
  const [messages, setMessage] = useState("");
  const [user, setUser] = useState(starterNames[Math.floor(Math.random()*starterNames.length)])
  const ws = useRef(null);

  useEffect(() => {
    //hardcoded backend wee
    ws.current = new WebSocket("ws://dave.quest:3000");
    ws.current.onopen = () => {
      ws.current.send(JSON.stringify({guy: user, kind: 0}))
    };
    return () => ws.current.close();
  }, []);

  useEffect(() => {
    ws.current.onmessage = (msg) => {
      setMessage(msg.data + "\n" + messages)
    }
  });

  function sendName(e) {
    e.preventDefault();
    if(user.length < 3) {
      return;
    }
    ws.current.send(JSON.stringify({guy: user, kind: 2}));
  }

  return (
    <>
      <h1> davechat </h1>

      <textarea rows="20" cols="40" value={messages} readOnly="true"/>

      <ChatBox socket={ws} name={user}/>
      <form onSubmit={sendName}>
        <input type="text" 
          rows={1} 
          cols={12} 
          maxLength={12} 
          value={user} 
          onChange={ e => setUser(e.target.value) }
        />
      </form>
    </>
  )
}

export default App
