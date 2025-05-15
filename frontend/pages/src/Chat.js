import React, { useEffect, useRef, useState } from "react";

function Chat({ roomId, username }) {
  const [messages, setMessages] = useState([]);
  const [text, setText] = useState("");

  const bottomRef = useRef(null);

  const scrollToBottom = () => {
    bottomRef.current?.scrollIntoView({ behavior: "smooth" ,block: "end"});
  };


  const fetchMessages = async () => {
    const token = sessionStorage.getItem("token");

    try {
      const res = await fetch(`http://localhost:8081/messages?room=${roomId}`, {
        headers: {
          "Authorization": `Bearer ${token}`,
        },
      });
      if (!res.ok) throw new Error("メッセージ取得に失敗");

      const data = await res.json();
      setMessages(data);
      scrollToBottom();
    } catch (err) {
      console.error("メッセージ取得エラー:", err);
    }
  };

  const handleSend = async () => {
    if (!text.trim()) return;

    const token = sessionStorage.getItem("token");

    try {
      const res = await fetch("http://localhost:8081/send", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          "Authorization": `Bearer ${token}`,
        },
        body: JSON.stringify({
          room_id: parseInt(roomId),
          text: text,
        }),
      });

      if (!res.ok) throw new Error("送信失敗");

      setText("");
      fetchMessages();
    } catch (err) {
      console.error("送信エラー:", err);
    }
  };

  useEffect(() => {
    fetchMessages();
    scrollToBottom();
  }, [roomId]);

  return (
    <div className="chat-container">
      <div className="chat-header">
        <h3>チャットルーム: {roomId}</h3>
      </div>

      <div className="chat-messages">
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
        <div ref={bottomRef} style={{ height: "32px" }} />
      </div>

      <div className="chat-input">
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
