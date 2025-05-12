import './App.css';
//useStateは変化するデータを持つ関数
import { useState } from 'react';

function App() {
  // 状態を定義する（データを持つ）
  // イベント処理などの関数を定義する

  //入力された内容をReactで保持するやつ
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState(''); 

  //登録を押したときにリクエストを送信するやつ
  const handleSignup = async () => {
    try {
      const response = await fetch('http://localhost:8081/signup', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({
          username: username,
          password: password
        })
      });

      if (response.ok) {
        alert('登録成功！');
      } else {
        const err = await response.text();
        alert(`登録失敗: ${err}`);
      }
    } catch (error) {
      console.error('通信エラー:', error);
      alert('通信エラーが発生しました、残念');
    }
  };

  return (
    //画面にどう表示するかを返している
    <div className="App">
      <h1>ChatApp</h1>
      <h2>新規登録</h2>

      
      <div className="form-row">
        <label>ユーザー名</label>
        <input 
          type="text"
          value={username}
          onChange={(e) => setUsername(e.target.value)}
        />
      </div>
      
      <div className="form-row">
        <label>パスワード</label>
        <input 
          type="password"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
        />
      </div>
      
      <p></p>
      <button onClick={handleSignup}>登録</button>
      <p></p>
      <h6>※ユーザーネームは日本語英語どちらでも大丈夫ですがパスワードは英語で入力してください</h6>
    </div>
  );
}

export default App;
