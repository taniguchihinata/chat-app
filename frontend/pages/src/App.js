//フロントエンドの機能の管理をしている

//必要なものをインポートしている
import React, { useEffect, useState } from "react";
import { Routes, Route, useLocation, useNavigate, Link } from "react-router-dom";
//Signupやチャットルームなどのルートをインポート
import Login from "./Login";
import Signup from "./Signup";
import UserList from "./UserList";
import ChatPage from "./ChatPage";
import RequireAuth from "./RequireAuth";
import CreateGroup from "./CreateGroup";


//アプリ全体を定義する関数型Reactコンポーネント
function App() {
  const location = useLocation();//現在のURLパスを取得
  const navigate = useNavigate();//ページ遷移をするときに必要

  const [username, setUsername] = useState(null);//今ログインしているユーザー名を保存
  const [reloadFlag, setReloadFlag] = useState(false);
  const showNav = location.pathname === '/' || location.pathname === '/signup';//ログイン画面かサインアップ画面の時のみログイン/サインアップボタンを表示

  const [notifications, setNotifications] = useState([]);
  const [showNotifications, setShowNotifications] = useState(false);
  const [unreadCount, setUnreadCount] = useState(0);

  //ログイン状態チェック
  useEffect(() => {
    const token = localStorage.getItem("token");//セッションからJWTトークンを取得
    if (!token) {//トークンがなかったらusernameをnullにする
      setUsername(null);
      return;
    }

    fetch("http://localhost:8081/me", {
      headers: {
        Authorization: `Bearer ${token}`,
      },
    })
      .then((res) => {
        if (!res.ok) throw new Error("認証エラー");
        return res.json();
      })
      .then((data) => {
        setUsername(data.username);
      })
      .catch((err) => {
        console.error("JWTからユーザー取得失敗:", err);
        localStorage.removeItem("token");
        setUsername(null);
        navigate("/");
      });
  }, [navigate]);

  //ログアウト処理
  //トークンを削除してログイン画面に戻る
  const handleLogout = () => {
    localStorage.removeItem("token");
    setUsername(null);
    navigate("/");
  };

  //通知一覧を取得
  useEffect(() => {
  if (!username) return;
  const fetchMentions = async () => {
    const token = localStorage.getItem("token");
    try {
      const res = await fetch("http://localhost:8081/mentions", {
        headers: { Authorization: `Bearer ${token}` },
      });
      const data = await res.json();
      if (Array.isArray(data)) {
        setNotifications(data);
        setUnreadCount(data.length);
      }
    } catch (err) {
      console.error("通知取得失敗:", err);
    }
  };

  fetchMentions();
}, [username]);

const handleNotificationClick = async (roomId, messageId) => {
  const token = localStorage.getItem("token");

  // ✅ 既読API呼び出し
  try {
    await fetch("http://localhost:8081/mentions/read", {
      method: "POST",
      headers: {
        "Authorization": `Bearer ${token}`,
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ message_id: messageId }),
    });
  } catch (err) {
    console.error("通知既読化失敗:", err);
  }
  navigate(`/chat/${roomId}?scrollTo=${messageId}`);
  setShowNotifications(false);
};

  //画面に表示される部分
  return (
    <div className="App">
      <nav style={{
        display: "flex",
        justifyContent: "space-between",
        alignItems: "center",
        padding: "10px",
        position: "relative"
      }}>
        {/* 左ブロック：ロゴやリンク */}
        {showNav ? (
          <div style={{ display: "flex", gap: "12px" }}>
            <Link to="/signup"><button>サインアップへ</button></Link>
            <Link to="/"><button>ログインへ</button></Link>
          </div>
        ) : (
          <div style={{ display: "flex", gap: "12px" }}>
            <button onClick={handleLogout}>ログアウト</button>
            <span style={{color: "white"}}>ログイン中: {username}</span>
          </div>
        )}

        {/* 右ブロック：通知アイコン（常に右端） */}
        {!showNav && username && (
          <div style={{ position: "absolute", right: "10px", top: "10px" }}>
            <button onClick={() => setShowNotifications(!showNotifications)} style={{ fontSize: "1.4em", position: "relative" }}>
              🔔
              {unreadCount > 0 && (
                <span style={{
                  position: "absolute", top: "-6px", right: "-6px",
                  background: "red", color: "white", borderRadius: "50%",
                  fontSize: "10px", padding: "2px 6px"
                }}>{unreadCount}</span>
              )}
            </button>

            {/* 通知ポップアップ */}
            {showNotifications && (
              <div style={{
                position: "absolute", top: "35px", right: "0",
                width: "300px", maxHeight: "300px", overflowY: "auto",
                background: "white", border: "1px solid #ccc", borderRadius: "8px",
                padding: "10px", zIndex: 1000
              }}>
                {notifications.length === 0 ? (
                  <div style={{ color: "#666" }}>メンションはありません</div>
                ) : (
                  notifications.map(n => (
                    <div
                      key={n.message_id}
                      onClick={() => handleNotificationClick(n.room_id, n.message_id)}
                      style={{
                        padding: "6px",
                        borderBottom: "1px solid #eee",
                        cursor: "pointer",
                        fontWeight: n.is_read ? "normal" : "bold", // 未読は太字に
                        color: "black"
                      }}
                    >
                      <strong>{n.sender_name}</strong>:<br />
                      {n.text ? (
                        <span>{n.text}</span>
                      ) : (
                        <span style={{ color: "#888" }}>[画像または空のメッセージ]</span>
                        )}
                        <br />
                        <small>{new Date(n.created_at).toLocaleString("ja-JP")}</small>
                      </div>

                  ))
                )}
              </div>
            )}
          </div>
        )}
      </nav>


      <div className="main-content">{/*それぞれのシステムへのルーティング*/}
        <Routes>
          <Route path="/" element={<Login />} />
          <Route path="/signup" element={<Signup />} />
          {/*RequireAuthでトークン無し不正ログインを防ぐ*/}
          
          <Route
            path="/chat/:roomId"
            element={
              <RequireAuth>
                <ChatPage username={username} />
              </RequireAuth>
            }
          />
          <Route
            path="/groups/create"
            element={
              <RequireAuth>
                <CreateGroup />
              </RequireAuth>
            }
          />
        
          <Route
            path="/chat"
            element={
              <RequireAuth>
                <UserList username={username} />
              </RequireAuth>
            }
          />
          <Route
            path="/chat"
            element={
              <RequireAuth>
                <UserList username={username} reloadFlag={reloadFlag} />
              </RequireAuth>
            }
          />
          <Route
            path="/chat/:roomId"
            element={
              <RequireAuth>
                <ChatPage username={username} onReadReaset={() => setReloadFlag((prev) => !prev)} />
              </RequireAuth>
            }
          />

        </Routes>
            
      </div>
    </div>
  );
}

//ほかのファイルからも利用できるようにする
export default App;
