//import './App.css';
import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import './FormContainer.css';

function Login() {
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [message, setMessage] = useState('');

  const navigate = useNavigate();

  const handleLogin = async () => {
    const response = await fetch('http://localhost:8081/login', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({ username, password })
    });

    if (response.ok) {
      setMessage('ログイン成功！');
      localStorage.setItem('username', username);//ログイン名を保存
      navigate('/users');//画面遷移
    } else {
      setMessage('ログイン失敗');
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
