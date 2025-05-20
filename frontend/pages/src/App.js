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

//アプリ全体を定義する関数型Reactコンポーネント
function App() {
  const location = useLocation();//現在のURLパスを取得
  const navigate = useNavigate();//ページ遷移をするときに必要

  const [username, setUsername] = useState(null);//今ログインしているユーザー名を保存

  const showNav = location.pathname === '/' || location.pathname === '/signup';//ログイン画面かサインアップ画面の時のみログイン/サインアップボタンを表示

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

  //画面に表示される部分
  return (
    <div className="App">
      <nav>
        {showNav ? (
          <>
            <Link to="/signup"><button>サインアップへ</button></Link>
            <Link to="/"><button>ログインへ</button></Link>
          </>
        ) : (
          username && (
            <>
              <span>ログイン中: {username}</span>
              <button onClick={handleLogout}>ログアウト</button>
            </>
          )
        )}
      </nav>

      <div className="main-content">{/*それぞれのシステムへのルーティング*/}
        <Routes>
          <Route path="/" element={<Login />} />
          <Route path="/signup" element={<Signup />} />
          {/*RequireAuthでトークン無し不正ログインを防ぐ*/}
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
    </div>
  );
}

//ほかのファイルからも利用できるようにする
export default App;
