import React, { useEffect, useRef, useState } from "react";
import { useInView } from "react-intersection-observer";

function MessageItem({ 
  msg,
  username,
  socketRef,
  roomId,
  readStatus,
  setReadStatus,
  sendWhenReady,
  readersByMessageId
}) {
  const isMine = msg.username === username;
  const [ref, inView] = useInView({ triggerOnce: true, threshold: 0.6 });

  useEffect(() => {
    if (!msg.id) return;

    if (inView && !isMine && !readStatus[msg.id]) {
      const token = localStorage.getItem("token");

      fetch("http://localhost:8081/read", {
        method: "POST",
        headers: {
          Authorization: `Bearer ${token}`,
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ message_id: msg.id }),
      }).catch((err) => console.error("既読登録失敗:", err));

      sendWhenReady({
        type: "read",
        room_id: parseInt(roomId),
        username,
        message_id: msg.id,
      });

      setReadStatus((prev) => ({
        ...prev,
        [msg.id]: true,
      }));
    }
  }, [inView, isMine, msg.id, roomId, username, sendWhenReady, readStatus, setReadStatus]);

  return (
    <div
      ref={ref}
      key={msg.id}
      className={isMine ? "my-message" : "other-message"}
      style={{
        textAlign: isMine ? "right" : "left",
        marginBottom: "4px",
        display: "flex",
        flexDirection: isMine ? "row-reverse" : "row",
        alignItems: "center",
      }}
    >
      {isMine && readersByMessageId[msg.id]?.length > 0 && (
        <span style={{ fontSize: "0.8rem", color: "gray", margin: "0 6px" }}>
          {readersByMessageId[msg.id].length === 1
            ? "既読"
            : `既読: ${readersByMessageId[msg.id].length}`}
        </span>
        )}

      <div className="message-content">
        <strong>{msg.username}: </strong>
        {msg.text}
        {msg.image && (
          <div style={{ marginTop: "4px" }}>
            <img 
            src={`http://localhost:8081${msg.image}`} 
            alt="添付画像" 
            style={{ maxWidth: "200px", norderRadius: "8px"}} />
          </div>
        )}
        <div style={{ fontSize: "0.7rem", color: "white", marginTop: "2px" }}>
          {new Date(msg.created_at).toLocaleString("ja-JP", {
            dateStyle: "short",
            timeStyle: "short",
          })}
        </div>
      </div>
    </div>
  );
}

function Chat({ roomId, username, onReadReaset }) {
  const [messages, setMessages] = useState([]);
  const [text, setText] = useState("");
  const [roomName, setRoomName] = useState("チャットルーム!");
  const [readStatus, setReadStatus] = useState({});
  const [readersByMessageId, setReadersByMessageId] = useState({});
  const [imageFile, setImageFile] = useState(null);

  const socketRef = useRef(null);
  const bottomRef = useRef(null);
  const fileInputRef = useRef(null);

  const handleFileChange = (e) => {
    setImageFile(e.target.files[0]);
    }
    useEffect(() => {
      const fetchRoomName = async () => {
        const token = localStorage.getItem("token");
        try {
          const res = await fetch(`http://localhost:8081/rooms/${roomId}`, {
            headers: { Authorization: `Bearer ${token}` },
          });
          const data = await res.json();
          if (data && data.name) {
            setRoomName(data.name);
          } else if (data && Array.isArray(data.users)) {
            const partnerName = data.users.find((user) => user !== username);
            setRoomName(partnerName || "相手未設定");
          }
        } catch (err) {
          console.error("ルーム名取得失敗:", err);
        }
      };

      if (roomId) {
        fetchRoomName();
      }
    }, [roomId, username]);

    useEffect(() => {
      if (!roomId || !username) return;
      const token = localStorage.getItem("token");

      const fetchReadStatus = async () => {
        try {
          const res = await fetch(`http://localhost:8081/read_status?room=${roomId}`, {
            headers: {
              Authorization: `Bearer ${token}`,
            },
          });
          const data = await res.json();
          if (Array.isArray(data)) {
            const readMap = {};
            data.forEach((id) => {
              readMap[id] = true;
            });
            setReadStatus((prev) => ({ ...prev, ...readMap }));
          }
        } catch (err) {
          console.error("既読状態の取得失敗:", err);
        }
      };

      const fetchReaders = async () => {
        try {
          const res = await fetch(`http://localhost:8081/read_status_full?room=${roomId}`, {
            headers: {
            Authorization: `Bearer ${token}` },
          });
          const data = await res.json();
          if (data && typeof data === "object") {
            setReadersByMessageId(data);
          }
        } catch (err) {
          console.error("既読ユーザー一覧取得失敗:", err);
        }
      };

    fetchReadStatus();
    fetchReaders();
  }, [roomId, username]);

  useEffect(() => {
    const markAllAsRead = async () => {
      const token = localStorage.getItem("token");
      try {
        await fetch(`http://localhost:8081/mark_all_read?room=${roomId}`, {
          method: "POST",
          headers: {
            Authorization: `Bearer ${token}`,
          },
        });
        if (onReadReaset) onReadReaset();
      } catch (err) {
        console.error("未読リセット失敗:", err);
      }
    };

    if (roomId) {
      markAllAsRead();
    }
  }, [roomId, onReadReaset]);

  useEffect(() => {
  return () => {
    // クリーンアップ時に「leave」通知
    const socket = socketRef.current;
    if (socket && socket.readyState === WebSocket.OPEN) {
      socket.send(JSON.stringify({
        type: "leave",
        room_id: parseInt(roomId),
        username,
      }));
    }
  };
}, [roomId, username]);


  const sendWhenReady = (messageObj) => {
    const socket = socketRef.current;
    if (!socket) return;

    if (socket.readyState === WebSocket.OPEN) {
      socket.send(JSON.stringify(messageObj));
    } else {
      const interval = setInterval(() => {
        if (socket.readyState === WebSocket.OPEN) {
          socket.send(JSON.stringify(messageObj));
          clearInterval(interval);
        }
      }, 100);
    }
  };

  useEffect(() => {
    const token = localStorage.getItem("token");
    if(!token){
      console.warn("トークンが存在しないため、WebSocket接続を中止します");
      return ;
    }
    const ws = new WebSocket(`ws://localhost:8081/ws?token=${token}`);
    socketRef.current = ws;

    ws.onopen = () => {
      sendWhenReady({ 
        type: "join",
        room_id: parseInt(roomId) });
    };

    ws.onmessage = (event) => {
      const msg = JSON.parse(event.data);
      msg.id = msg.id ?? msg.message_id;

      if (msg.type === "read"){
        setReadStatus((prev) => ({
          ...prev,
          [msg.message_id]: true,
        }));

        setReadersByMessageId((prev) => {
          const readers = prev[msg.message_id] || [];
          if (!readers.includes(msg.username)) {
            return {
              ...prev,
              [msg.message_id]: [...readers, msg.username],
            };
          }
          return prev;
        });

        return;
      }

      if (msg.type === "leave"){
        return;
      }

      if (!msg.created_at){
        msg.created_at = new Date().toISOString();
      }

      setMessages((prev) => 
        Array.isArray(prev) ? [...prev, msg] : [msg]
      );
    };

    return () => {
    if (ws.readyState === WebSocket.OPEN) {
      ws.send(JSON.stringify({
        type: "leave",
        room_id: parseInt(roomId),
        username,
      }));
      ws.close();
    } else {
      console.warn("WebSocket未接続のため、leave/send/closeをスキップします。");
    }
  };
  }, [roomId]);

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

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: "auto", block: "end" });
  }, [messages]);

  const handleSend = async () => {
    if (!text.trim() && !imageFile) return;

    let imageUrl = "";
    if (imageFile) {
      const formData = new FormData();
      formData.append("image", imageFile);
      try {
        const token = localStorage.getItem("token");
        const res = await fetch("http://localhost:8081/upload", {
          method: "POST",
          headers: { Authorization: `Bearer ${token}` },
          body: formData,
        });
        if (res.ok) {
          const data = await res.json();
          imageUrl = data.file_url;
        } else {
          alert("画像アップロードに失敗しました");
        }
      } catch (err) {
        console.error("画像送信エラー:", err);
      }
    }

    sendWhenReady({
      type: "message",
      room_id: parseInt(roomId),
      text: text.trim(),
      image: imageUrl,
    });
    setText("");
    setImageFile(null);
    if (fileInputRef.current){
      fileInputRef.current.value = "";
    }
  };


  return (
    <div className="chat-container">
      <div className="chat-header">
        <h3>チャット for {roomName}</h3>
      </div>

      <div 
        className="chat-messages" 
        style={{ overflowY: "auto", maxHeight: "60vh", padding: "0 10px" }}
      >
        {messages?.map((msg) => (
          <MessageItem
            key={`${msg.id}-${readStatus[msg.id] ? 'read' : 'unread'}`}
            msg={msg}
            username={username}
            socketRef={socketRef}
            roomId={roomId}
            readStatus={readStatus}
            setReadStatus={setReadStatus}
            sendWhenReady={sendWhenReady}
            readersByMessageId={readersByMessageId}
          />
        ))}
        <div ref={bottomRef} style={{ height: "0", margin: 0, padding: 0 }} />
      </div>

      <div style={{ display: "flex", alignItems: "center", marginBottom: "8px" }}>
        <input type="file" accept="image/*" onChange={handleFileChange} />
      </div>

      <div className="chat-input" style={{ marginTop: "10px"}}>
        <textarea
          value={text}
          onChange={(e) => setText(e.target.value)}
          onKeyDown={(e) => {
            if (e.key === "Enter" && !e.shiftKey) {
              e.preventDefault();
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
