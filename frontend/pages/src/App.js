import { Routes, Route, Link, useLocation, useNavigate } from 'react-router-dom';
import { useEffect, useState } from "react";
import Signup from './Signup';
import Login from './Login';
import UserList from './UserList';
import ChatPage from './ChatPage';

function App() {
  const location = useLocation();
  const navigate = useNavigate();
  const [username, setUsername] = useState(null);

  const showNav = location.pathname === '/' || location.pathname === '/signup';

  useEffect(() => {
    const token = localStorage.getItem("token");
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
        localStorage.removeItem("token");
        setUsername(null);
        navigate("/"); // トークン無効ならログイン画面へ
      });
  }, [navigate]);

  return (
    <div className="App">
      {showNav && (
        <nav>
          <Link to="/signup"><button>サインアップへ</button></Link>
          <Link to="/"><button>ログインへ</button></Link>
        </nav>
      )}

      {!showNav && username && (
        <nav>
          <span>ログイン中: {username}</span>
          <button onClick={() => {
            localStorage.removeItem("token");
            setUsername(null);
            navigate("/");
          }}>ログアウト</button>
        </nav>
      )}

      <Routes>
        <Route path="/" element={<Login />} />
        <Route path="/signup" element={<Signup />} />
        <Route path="/users" element={<UserList username={username} />} />
        <Route path="/chat/:roomId" element={<ChatPage username={username} />} />
      </Routes>
    </div>
  );
}

export default App;
