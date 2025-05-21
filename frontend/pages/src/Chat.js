import React, { useEffect, useRef, useState, } from "react";
import { useInView } from "react-intersection-observer";


function MessageItem({ msg, username, socketRef, roomId, readStatus, setReadStatus, sendWhenReady }) {
  const isMine = msg.username === username;
  const [ref, inView] = useInView({
    triggerOnce: true,
    threshold: 0.6,
  });

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
      {isMine && readStatus[msg.id] && (
        <span style={{ fontSize: "0.8rem", color: "gray", margin: "0 6px" }}>
          既読
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

  const socketRef = useRef(null)
  const bottomRef = useRef(null);

  // ✅ readStatus を roomId 有効時に復元
  useEffect(() => {
    if (!roomId || !username) return;
    try {
      const saved = localStorage.getItem(`readStatus-${roomId}-${username}`);
      setReadStatus(saved ? JSON.parse(saved) : {});
    } catch {
      setReadStatus({});
    }
  }, [roomId, username]);

  //既読状態の永続化
  //readStatusの状態をlocaStorageに保存する
  useEffect(() => {
    if (roomId && username) {
      localStorage.setItem(`readStatus-${roomId}-${username}`, JSON.stringify(readStatus));
    }  
  }, [readStatus, roomId, username]);

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

  // ルーム参加時に既読状態取得
useEffect(() => {
  if (!roomId || !username) return;
  const fetchReadStatus = async () => {
    const token = localStorage.getItem("token");
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

  fetchReadStatus();
}, [roomId, username]);


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
        console.log("既読通知受信:",msg);

        //自分が送信したメッセージに対して既読マークを表示する
        setReadStatus((prev) => ({
          ...prev,
          [msg.message_id]: true,
        }));
        return;
      }

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

  // 新しいメッセージが届いたときに既読通知を送信
  useEffect(() => {
    if (!messages || messages.length === 0) return;
    const lastMsg = messages[messages.length - 1];

    const isMine = lastMsg.username === username;
    if (!isMine && !readStatus[lastMsg.id]) {
      const token = localStorage.getItem("token");

      fetch("http://localhost:8081/read", {
        method: "POST",
        headers: {
          Authorization: `Bearer ${token}`,
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ message_id: lastMsg.id }),
      }).catch((err) => console.error("既読登録失敗:", err));

      sendWhenReady({
        type: "read",
        room_id: parseInt(roomId),
        username,
        message_id: lastMsg.id,
      });
    }
  }, [messages, username, roomId, readStatus]);

  //ルーム情報取得
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

    sendWhenReady({
      type: "message",
      room_id: parseInt(roomId),
      text: text.trim(),
    });

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
        <h3>チャット for {roomName}</h3>
      </div>

      <div 
        className="chat-messages" 
        style={{ overflowY: "auto", maxHeight: "60vh", padding: "0 10px" }}
      >
        {messages?.map((msg) => (
          <MessageItem
            key={msg.id}
            msg={msg}
            username={username}
            socketRef={socketRef}
            roomId={roomId}
            readStatus={readStatus}
            setReadStatus={setReadStatus}
            sendWhenReady={sendWhenReady}
          />
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
