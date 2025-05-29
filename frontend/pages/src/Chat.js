//チャットルームのフロントエンド
import React, { useEffect, useRef, useState } from "react";
import { useInView } from "react-intersection-observer";
import { useLocation } from "react-router-dom";


function MessageItem({ 
  msg,
  username,
  socketRef,
  roomId,
  readStatus,
  setReadStatus,
  sendWhenReady,
  readersByMessageId,
  scrollRefs,
  onUndo,
  onDelete,
  isAtBottom,
  readRequestedRef
}) {
  const isMine = msg.username === username;
  const localRef = useRef(null);
  const [inViewRef, inView] = useInView({ triggerOnce: true, threshold: 0.6});
  const [hovered, setHovered] = useState(false);

  useEffect(() => {
    if (msg.id && scrollRefs?.current) {
      scrollRefs.current[msg.id] = localRef.current;
    }
  }, [msg.id, scrollRefs]);

  useEffect(() => {
    if (!msg.id || isMine) return;

    if (readStatus[msg.id] || readRequestedRef.current.has(msg.id)) return;

    //if (readRequestedRef.current.has(msg.id)) return;

    if (inView) {
      readRequestedRef.current.add(msg.id);
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
  }, [inView, isMine, msg.id, roomId, username, sendWhenReady, readStatus, setReadStatus, readRequestedRef]);

  return (
    <div
      ref={(el) => {
        inViewRef(el);
        localRef.current = el;
      }}
      key={msg.id}
      onMouseEnter={() => setHovered(true)}
      onMouseLeave={() => setHovered(false)}
      style={{
        display: "flex",
        justifyContent: isMine ? "flex-end" : "flex-start",
        marginBottom: "8px",
      }}
    >
      <div
        style={{
          display: "flex",
          flexDirection: "row",
          alignItems: "center",
          background: isMine ? "#ffe8cc" : "#f0f0f0",
          borderRadius: "8px",
          padding: "6px 10px",
          maxWidth: "80%",
        }}
      >
        {/* メッセージ内容 */}
        <div>
          {msg.deleted ? (
            <em style={{ color: "gray" }}>このメッセージは削除されました</em>
          ) : (
            <>
              <strong>{msg.username}: </strong>
              {msg.text}
              {msg.image && (
                <div style={{ marginTop: "4px" }}>
                  <img
                    src={`http://localhost:8081${msg.image}`}
                    alt="添付画像"
                    onLoad={() => {
                      if (!isAtBottom) return;
                      const el = document.getElementById("bottom-ref");
                      el?.scrollIntoView({ behavior: "auto", block: "end" });
                    }}
                    style={{
                      width: msg.image.startsWith("/stamps/") ? "80px" : "auto",
                      height: msg.image.startsWith("/stamps/") ? "80px" : "auto",
                      maxWidth: msg.image.startsWith("/stamps/") ? "none" : "200px",
                      borderRadius: msg.image.startsWith("/stamps/") ? "0px" : "8px",
                      backgroundColor: msg.image.startsWith("/stamps/") ? "transparent" : undefined,
                      objectFit: "contain",
                    }}
                  />

                </div>
              )}
            </>
          )}
          <div style={{ fontSize: "0.7rem", color: "gray", marginTop: "2px" }}>
            {new Date(msg.created_at).toLocaleString("ja-JP", {
              dateStyle: "short",
              timeStyle: "short",
            })}
            {isMine && readersByMessageId[msg.id]?.length > 0 && (
              <span style={{ marginLeft: "8px", color: "gray" }}>
                {readersByMessageId[msg.id].length === 1
                  ? "既読"
                  : `既読: ${readersByMessageId[msg.id].length}`}
              </span>
            )}
          </div>
        </div>

        {/* 操作ボタン */}
        {isMine && hovered && !msg.deleted && (
          <div style={{ marginLeft: "10px", display: "flex", gap: "6px" }}>
            {Date.now() - new Date(msg.created_at).getTime() < 60 * 1000 && (
              <button
                onClick={() => onUndo(msg.id)}
                style={{
                  fontSize: "0.7rem",
                  background: "none",
                  border: "none",
                  color: "blue",
                  cursor: "pointer",
                }}
              >
                取り消し
              </button>
            )}
            <button
              onClick={() => onDelete(msg.id)}
              style={{
                fontSize: "0.7rem",
                background: "none",
                border: "none",
                color: "red",
                cursor: "pointer",
              }}
            >
              削除
            </button>
          </div>
        )}
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
  
  const location = useLocation();
  const searchParams = new URLSearchParams(location.search);
  const scrollToId = searchParams.get("scrollTo"); // 例: "123"
  const scrollRefs = useRef({});
  const readRequestedRef = useRef(new Set());
  const socketRef = useRef(null);
  const bottomRef = useRef(null);
  const fileInputRef = useRef(null);
  const STAMP_LIST = [
    "/stamps/smile.png",
    "/stamps/angry.png",
    "/stamps/love.png",
    "/stamps/laugh.png",
  ];

  const handleUndo = async (messageId) => {
    const token = localStorage.getItem("token");
    try {
      const res = await fetch(`http://localhost:8081/messages/${messageId}`, {
        method: "DELETE",
        headers: { Authorization: `Bearer ${token}` },
      });
      if (!res.ok) throw new Error("削除失敗");

      sendWhenReady({
        type: "delete",
        room_id: parseInt(roomId),
        message_id: messageId,
        text: "hard"
      });

      setMessages((prev) => prev.filter((msg) => msg.id !== messageId));
    } catch (err) {
      console.error("送信取り消しエラー:", err);
    }
  };

  const handleDelete = async (messageId) => {
    const token = localStorage.getItem("token");

    // 対象メッセージを取得
    const targetMsg = messages.find((m) => m.id === messageId);
    if (!targetMsg) return;

    const createdTime = new Date(targetMsg.created_at).getTime();
    const now = Date.now();
    const within60s = now - createdTime < 60 * 1000;

    try {
      if (within60s) {
        // 送信取り消し（物理削除）
        const res = await fetch(`http://localhost:8081/messages/${messageId}`, {
          method: "DELETE",
          headers: { Authorization: `Bearer ${token}` },
        });
        if (!res.ok) throw new Error("送信取り消し失敗");

        sendWhenReady({
          type: "delete",
          room_id: parseInt(roomId),
          message_id: messageId,
          text: "hard,"
        });

        setMessages((prev) => prev.filter((msg) => msg.id !== messageId));
      } else {
        // 通常の削除（論理削除）
        const res = await fetch(`http://localhost:8081/messages/${messageId}`, {
          method: "PATCH",
          headers: {
            Authorization: `Bearer ${token}`,
            "Content-Type": "application/json",
          },
          body: JSON.stringify({ deleted: true }),
        });
        if (!res.ok) throw new Error("削除失敗");

        sendWhenReady({
          type: "delete",
          room_id: parseInt(roomId),
          message_id: messageId,
        });

        setMessages((prev) =>
          prev.map((msg) =>
            msg.id === messageId ? { ...msg, deleted: true } : msg
          )
        );
      }
    } catch (err) {
      console.error("削除処理エラー:", err);
    }
  };


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
      console.log("受信メッセージ:", msg);
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

      if (msg.type === "delete") {
        if (msg.hard_delete) {
          setMessages((prev) => prev.filter((m) => m.id !== msg.message_id));
        } else {
          setMessages((prev) => prev.map((m) => m.id === msg.message_id ? { ...m, deleted: true } : m) )
        };
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
  }, [roomId, username]);

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
    if (scrollToId && scrollRefs.current[scrollToId]) {
      setTimeout(() => {
        scrollRefs.current[scrollToId]?.scrollIntoView({
          behavior: "auto",
          block: "center",
        });
      }, 100);
    }
  }, [messages, scrollToId]);

  const prevMessageCountRef = useRef(0);

  useEffect(() => {

    if (messages.length > prevMessageCountRef.current) {
      bottomRef.current?.scrollIntoView({ behavior: "auto", block: "end" });
    }
  prevMessageCountRef.current = messages.length;  }, [messages]);

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

  const handleStampSend = (stampUrl) => {
    sendWhenReady({
      type: "message",
      room_id: parseInt(roomId),
      text: "",           // スタンプなのでテキストなし
      image: stampUrl,    // スタンプの画像URLを image として送信
    });

  

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
            scrollRefs={scrollRefs}
            onUndo={handleUndo}
            onDelete={handleDelete}
            readRequestedRef={readRequestedRef}
          />
        ))}
        <div ref={bottomRef} id="bottom-ref" style={{ height: "0", margin: 0, padding: 0 }} />
      </div>

      <div style={{ display: "flex", alignItems: "center", marginBottom: "8px" }}>
        <input 
          type="file" 
          accept="image/*" 
          ref={fileInputRef}
          onChange={handleFileChange} />
      </div>

      <div style={{ display: "flex", gap: "10px", marginBottom: "8px" }}>
        {STAMP_LIST.map((url, idx) => (
          <img
            key={idx}
            src={`http://localhost:8081${url}`}
            alt={`stamp-${idx}`}
            onClick={() => handleStampSend(url)}
            style={{ width: "25px", height: "25px", cursor: "pointer" }}
          />
        ))}
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
