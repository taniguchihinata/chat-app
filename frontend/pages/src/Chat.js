import React, { useEffect, useState } from "react";

function Chat({ roomId, username }) {
  const [messages, setMessages] = useState([]);
  const [text, setText] = useState("");

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
  }, [roomId]);

  return (
    <div>
      <h3>チャットルーム: {roomId}</h3>
      <div
        style={{
          border: "1px solid #ccc",
          height: "300px",
          overflowY: "scroll",
          padding: "8px",
          marginBottom: "8px",
        }}
      >
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
      </div>
      <input
        type="text"
        value={text}
        onChange={(e) => setText(e.target.value)}
        style={{ width: "80%", marginRight: "8px" }}
      />
      <button onClick={handleSend}>送信</button>
    </div>
  );
}

export default Chat;
