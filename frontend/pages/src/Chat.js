import { useParams } from 'react-router-dom';
import { useState, useEffect } from 'react';

function Chat() {
  const { username } = useParams(); // チャット相手のユーザー名
  const currentUser = localStorage.getItem('username'); // 自分のユーザー名

  const [messages, setMessages] = useState([]);
  const [text, setText] = useState('');

  // UTF-8 安全な Base64 エンコード関数
  function encodeBase64Unicode(str) {
    const utf8Bytes = new TextEncoder().encode(str);
    return btoa(String.fromCharCode(...utf8Bytes));
  }

  const encodedUser = encodeBase64Unicode(currentUser); // ← fetch用ヘッダーに使う

  // メッセージ一覧取得
  useEffect(() => {
    const fetchMessages = async () => {
      try {
        const res = await fetch(`http://localhost:8081/messages?with=${username}`, {
          headers: {
            'X-User': encodedUser
          }
        });
        if (!res.ok) throw new Error('メッセージ取得に失敗');
        const data = await res.json();
        setMessages(data);
      } catch (err) {
        console.error('メッセージ取得エラー:', err);
      }
    };

    fetchMessages();
  }, [username, encodedUser]);

  // メッセージ送信処理
  const handleSend = async () => {
    if (text.trim() === '') return;

    try {
      const res = await fetch('http://localhost:8081/send', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'X-User': encodedUser
        },
        body: JSON.stringify({
          to: username,
          content: text
        })
      });

      if (res.ok) {
        setText('');
        // 再取得（再度 fetchMessages を inline 定義）
        const newRes = await fetch(`http://localhost:8081/messages?with=${username}`, {
          headers: {
            'X-User': encodedUser
          }
        });
        if (newRes.ok) {
          const newData = await newRes.json();
          setMessages(newData);
        }
      } else {
        const errText = await res.text();
        throw new Error(errText || '送信失敗');
      }
    } catch (err) {
      console.error('送信エラー:', err);
    }
  };

  return (
    <div>
      <h2>チャット相手: {username}</h2>

      <div style={{ border: '1px solid #ccc', padding: '10px', minHeight: '200px' }}>
        {messages.map((msg, index) => (
          <p
            key={index}
            style={{ textAlign: msg.from_name === currentUser ? 'right' : 'left' }}
          >
            <strong>{msg.from_name}:</strong> {msg.content}
          </p>
        ))}
      </div>

      <div style={{ marginTop: '10px' }}>
        <input
          value={text}
          onChange={(e) => setText(e.target.value)}
          placeholder="メッセージを入力"
          style={{ width: '70%' }}
        />
        <button onClick={handleSend}>送信</button>
      </div>
    </div>
  );
}

export default Chat;
