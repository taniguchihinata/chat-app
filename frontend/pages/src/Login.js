import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import './FormContainer.css';

//ログインフォームの見た目と機能を定義
function Login() {
  //ユーザー情報、エラーメッセージを格納
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [message, setMessage] = useState('');

  const navigate = useNavigate();

  //ボタンを押したときに動くやつ。
  const handleLogin = async () => {
    if (!username.trimStart() || !password.trim()) return;

    try {
      //APIにPOSTを送信 
      const response = await fetch('http://localhost:8081/login', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ username, password }),
      });

      //HTTPステータスが200番台でなければ失敗と判断
      if (!response.ok) {
        const errorText = await response.text(); // プレーンテキスト読み取り
        setMessage(`ログイン失敗: ${errorText}`);
        return;
      }

      //レスポンスをJSONとして受け取りtokenとusrenameをseseionStageに保存
      //ページ遷移の認証で利用する
      const responseData = await response.json(); // 🔧 変数名を重複させない
      localStorage.setItem('token', responseData.token); // JWT保存
      localStorage.setItem('username', username);        // ユーザー名も保存

      
      setMessage('ログイン成功！');
      navigate('/users');
    } catch (err) {
      console.error('ログイン中にエラー:', err);
      setMessage('ログイン中にエラーが発生しました');
    }
  };

  return (
    <div className="form-container">
      <h1>ChatApp</h1>
      <h2>ログイン</h2>
      <div>
        <label>ユーザー名</label>
        <input
          value={username}
          onChange={(e) => setUsername(e.target.value)}
        />
      </div>

      <div>
        <label>パスワード</label>
        <input
          type="password"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
        />
      </div>

      <p></p>
      <button onClick={handleLogin}>ログイン</button>
      <p>{message}</p>
    </div>
  );
}

export default Login;