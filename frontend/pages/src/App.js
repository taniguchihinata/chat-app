import React, { useEffect, useState } from "react";
import { Routes, Route, useLocation, useNavigate, Link } from "react-router-dom";

import Login from "./Login";
import Signup from "./Signup";
import UserList from "./UserList";
import ChatPage from "./ChatPage";
import RequireAuth from "./RequireAuth";

function App() {
  const location = useLocation();
  const navigate = useNavigate();
  const [username, setUsername] = useState(null);

  const showNav = location.pathname === '/' || location.pathname === '/signup';

  useEffect(() => {
    const token = sessionStorage.getItem("token");
    if (!token) {
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
        sessionStorage.removeItem("token");
        setUsername(null);
        navigate("/");
      });
  }, [navigate]);

  const handleLogout = () => {
    sessionStorage.removeItem("token");
    setUsername(null);
    navigate("/");
  };

  return (
    <div className="App">
      {showNav ? (
        <nav>
          <Link to="/signup"><button>サインアップへ</button></Link>
          <Link to="/"><button>ログインへ</button></Link>
        </nav>
      ) : (
        username && (
          <nav>
            <span>ログイン中: {username}</span>
            <button onClick={handleLogout}>ログアウト</button>
          </nav>
        )
      )}

      <Routes>
        <Route path="/" element={<Login />} />
        <Route path="/signup" element={<Signup />} />
        <Route
          path="/users"
          element={
            <RequireAuth>
              <UserList username={username} />
            </RequireAuth>
          }
        />
        <Route
          path="/chat/:roomId"
          element={
            <RequireAuth>
              <ChatPage username={username} />
            </RequireAuth>
          }
        />
      </Routes>
    </div>
  );
}

export default App;

