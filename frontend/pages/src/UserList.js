import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';

function UserList() {
  const [users, setUsers] = useState([]);
  const [error, setError] = useState('');
  const navigate = useNavigate();

  useEffect(() => {
    fetch('http://localhost:8081/users')
      .then((res) => {
        if (!res.ok) throw new Error('ユーザー取得に失敗しました');
        return res.json();
      })
      .then((data) => {
        const currentUser = localStorage.getItem('username');//現在のユーザー名
        const filtered = data.filter((user) => user.username !== currentUser);//除外
        setUsers(filtered);
      })
      .catch((err) => setError(err.message));
  }, []);

  const handleLogout = () => {
    localStorage.removeItem('username');//ログイン情報削除
    navigate('/');//ログイン画面に戻る
  }

  return (
    <div>
      
      <h2>ユーザー一覧</h2>
      {error && <p style={{ color: 'red' }}>{error}</p>}
      <ul>
        {users.map((user) => (
          <li key={user.id}>{user.username}</li>
        ))}
      </ul>
      <button onClick={handleLogout}>ログアウト</button>
    </div>
  );
}

export default UserList;
