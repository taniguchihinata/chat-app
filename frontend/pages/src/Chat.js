import React, { useEffect, useRef, useState, } from "react";

function Chat({ roomId, username }) {
  const [messages, setMessages] = useState([]);
  const [text, setText] = useState("");
  const [roomName, setRoomName] = useState("チャットルーム");

  const socketRef = useRef(null)
  const bottomRef = useRef(null);

  useEffect(() => {
    const fetchMessages = async () => {
      const token = localStorage.getItem("token");
      try {
        const res = await fetch(`http://localhost:8081/messages?room=${roomId}`, {
          headers: {
            Authorization: `Bearer ${token}`,
          },
        });
        if (!res.ok) throw new Error("メッセージ取得に失敗");
        const data = await res.json();
        setMessages(Array.isArray(data) ? data : []);
      } catch (err) {
        console.error("過去メッセージ取得エラー:", err);
      }
    };

    fetchMessages();
  }, [roomId]);

  //WebSocket接続とメッセージ受信
  useEffect(() => {
    const token = localStorage.getItem("token");
    const ws = new WebSocket(`ws://localhost:8081/ws?token=${token}`);
    socketRef.current = ws;

    ws.onopen = () => {
      // 初期メッセージでルームID送信
      ws.send(JSON.stringify({
        room_id: parseInt(roomId),
      }));
    };

    ws.onmessage = (event) => {
      const msg = JSON.parse(event.data);
      setMessages((prev) => 
        Array.isArray(prev) ? [...prev, msg] : [msg]
      );
    };

    ws.onerror = (err) => {
      console.debug("WebSocket一時的な接続エラー（開発中などでよくある）:", err);
    }

    ws.onclose = () => {
      console.log("WebSocket切断");
    };

    return () => {
      ws.close();
    };
  }, [roomId]);

  useEffect(() => {
  const fetchRoomInfo = async () => {
    const token = localStorage.getItem("token");
    try {
      const res = await fetch(`http://localhost:8081/rooms/${roomId}`, {
        headers: { Authorization: `Bearer ${token}` },
      });
      const data = await res.json();
      if (data.type === "group") {
        setRoomName(data.name);
      } else {
        const other = data.users.find((u) => u !== username);
        setRoomName(other);
      }
    } catch {
      setRoomName("チャットルーム");
    }
  };
  fetchRoomInfo();
}, [roomId, username]);


  //メッセージ送信
  const handleSend = () => {
    if (!text.trim()) return;

    socketRef.current?.send(JSON.stringify({
      room_id: parseInt(roomId),
      text: text.trim(),
    }));

    setText("");
  };

  //初回レンダリング時の処理
  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: "smooth", block: "end"});
  }, [messages]);

  //UI
  return (
    <div className="chat-container">
      <div className="chat-header">
        <h3>チャット: {roomName}</h3>
      </div>

      <div className="chat-messages" style={{ overflowY: "auto", maxHeight: "60vh", padding: "0 10px" }}>
        {messages?.map((msg) => (
          <div
            key={msg.id} 
            style={{
              textAlign: msg.username === username ? "right" : "left",
              marginBottom: "4px",
            }}
          >
            <strong>{msg.username}: </strong>
            {msg.text}
          </div>
        ))}
        <div ref={bottomRef} style={{ height: "0" ,margin: 0, padding: 0}} />
      </div>

      <div className="chat-input" style={{ marginTop: "10px"}}>
        <textarea
          value={text}
          onChange={(e) => setText(e.target.value)}
          onKeyDown={(e) => {
            if (e.key === "Enter") {
              if (e.shiftKey) return; // 改行を許可
              e.preventDefault();     // 通常送信
              handleSend();
            }
          }}
          rows={3}
          style={{ 
            width: "80%",
            marginRight: "8px",
            resize: "none",
            lineHeight: "1.4em",
            padding: "10px",
            boxSizing: "border-box",
            
          }}
        />
        <button onClick={handleSend}>送信</button>
      </div>
    </div>
  );
}

export default Chat;
