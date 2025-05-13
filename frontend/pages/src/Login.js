import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import './FormContainer.css';

function Login() {
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [message, setMessage] = useState('');

  const navigate = useNavigate();

  const handleLogin = async () => {
    try {
      const response = await fetch('http://localhost:8081/login', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ username, password }),
      });

      if (!response.ok) {
        const errorText = await response.text(); // プレーンテキスト読み取り
        setMessage(`ログイン失敗: ${errorText}`);
        return;
      }

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