// Chat.js（readersByMessageId を /read_status_full で初期化する版）
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
    if (inView && !isMine && !readStatus[msg.id]) {
      const token = localStorage.getItem("token");

      // ✅ /read API呼び出し
      fetch("http://localhost:8081/read", {
        method: "POST",
        headers: {
          Authorization: `Bearer ${token}`,
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ message_id: msg.id }),
      }).catch((err) => console.error("既読登録失敗:", err));

      // ✅ WebSocketで既読通知送信
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
      {isMine && readersByMessageId[msg.id]?.length > 0  && (
        <span style={{ fontSize: "0.8rem", color: "gray", margin: "0 6px" }}>
          既読: {readersByMessageId[msg.id].join(",")}
        </span>
      )}
      <div className="message-content">
        <strong>{msg.username}: </strong>
        {msg.text}
      </div>
    </div>
  );
}

function Chat({ roomId, username }) {
  const [messages, setMessages] = useState([]);
  const [text, setText] = useState("");
  const [roomName, setRoomName] = useState("チャットルーム");
  const [readStatus, setReadStatus] = useState({});
  const [readersByMessageId, setReadersByMessageId] = useState({});

  const socketRef = useRef(null)
  const bottomRef = useRef(null);

  // ルーム参加時に既読状態取得
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

  //WebSocket接続とメッセージ受信
  useEffect(() => {
    const token = localStorage.getItem("token");
    const ws = new WebSocket(`ws://localhost:8081/ws?token=${token}`);
    socketRef.current = ws;

    ws.onopen = () => {
      sendWhenReady({ 
        type: "join",
        room_id: parseInt(roomId) });
    };

    ws.onmessage = (event) => {
      const msg = JSON.parse(event.data);

      if (msg.type === "read"){
        //自分が送信したメッセージに対して既読マークを表示する
        setReadStatus((prev) => ({
          ...prev,
          [msg.message_id]: true,
        }));

        // 読んだユーザー一覧に追加
        setReadersByMessageId((prev) => {
          const readers = prev[msg.message_id] || [];
          // 重複チェック
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

      setMessages((prev) => 
        Array.isArray(prev) ? [...prev, msg] : [msg]
      );
    };

    ws.onclose = () => {
      console.log("WebSocket切断");
    };

    return () => {
      ws.close();
    };
  }, [roomId]);

  //メッセージの取得
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

  //メッセージ送信
  const handleSend = () => {
    if (!text.trim()) return;

    sendWhenReady({
      type: "message",
      room_id: parseInt(roomId),
      text: text.trim(),
    });
    setText("");
  };

  //UI
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
        <div ref={bottomRef} style={{ height: "0" ,margin: 0, padding: 0}} />
      </div>

      <div className="chat-input" style={{ marginTop: "10px"}}>
        <textarea
          value={text}
          onChange={(e) => setText(e.target.value)}
          onKeyDown={(e) => {
            if (e.key === "Enter" && !e.shiftKey) {
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
