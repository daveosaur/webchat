import { useState, useRef, useEffect } from 'react'
import ChatBox from './chatbox.jsx'

function App() {
  const [messages, setMessage] = useState("");
  const [logged, setLogged] = useState(false)
  // const [user, setUser] = useState(starterNames[Math.floor(Math.random()*starterNames.length)])
  const [user, setUser] = useState("")
  const ws = useRef(null);
  const chatW = useRef(null);

  useEffect(() => {
    //hardcoded backend wee
    ws.current = new WebSocket("ws://dave.quest:3000");
    // ws.current.onopen = () => {
    //   // ws.current.send(JSON.stringify({guy: user, kind: 0}))
    // };
    return () => ws.current.close();
  }, []);

  useEffect(() => {
    if(chatW.current) {
      chatW.current.scrollTop = chatW.current.scrollHeight;
    }
    ws.current.onmessage = (msg) => {
      const parsed = JSON.parse(msg.data)
      console.log(parsed)
      switch(parsed.kind) {
        case 0: //connect
          setUser(parsed.guy)
          setLogged(true)
          break;
        case 1: //message
          const newMsg = `<${parsed.guy}> ${parsed.msg}`
          setMessage(messages + "\n" + newMsg)
          break;
        case 2, 3: //server message
          setMessage(messages + "\n" + parsed.msg)
          break;
      }
    }
  }, [messages]);

  function sendName(e) {
    e.preventDefault();
    if(user.length < 3) {
      return;
    }
    ws.current.send(JSON.stringify({guy: user, kind: 0}));
  }
  function rename(e) {
    e.preventDefault();
    ws.current.send(JSON.stringify({guy: user, kind: 2}));
  }

  return (
    <div align="center">
      <h1> davechat </h1>

      {logged ? (
        <>
          <textarea 
            ref={chatW}
            rows="20" 
            cols="40" 
            value={messages} 
            readOnly="true"/>
          <ChatBox socket={ws} name={user}/>
          <form onSubmit={rename}>
            <input type="text" 
              rows={1} 
              cols={12} 
              maxLength={12} 
              value={user} 
              onChange={e => setUser(e.target.value)}/>
          </form>
        </> ) : (
        <>
          <p> set ur name :))) </p>
          <form onSubmit={sendName}>
            <input type="text" 
              rows={1} 
              cols={12} 
              maxLength={12} 
              value={user} 
              onChange={ e => setUser(e.target.value) } />
            {(user.length < 3) ?
            (<button disabled={true}>enter a name pls</button>) :
            (<button>continue</button>)
          }
          </form>
        </> )}
    </div>
  )
}

export default App

