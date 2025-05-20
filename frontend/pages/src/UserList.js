import React, { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";

function UserList() {
  const [users, setUsers] = useState([]);
  const navigate = useNavigate();

  // ユーザーリスト一覧を取得
  useEffect(() => {
    const fetchUsers = async () => {
      const myUsername = localStorage.getItem("username");

      const res = await fetch("http://localhost:8081/users");
      const data = await res.json();

      // 自分以外のユーザーだけ表示
      const filtered = data.filter((user) => user.username !== myUsername);
      setUsers(filtered);
    };

    fetchUsers();
  }, []);

  // チャットルーム作成のハンドラ
  const handleClick = async (partnerUsername) => {
    const token = localStorage.getItem("token");

    try {
      const res = await fetch("http://localhost:8081/rooms", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          "Authorization": `Bearer ${token}`,
        },
        body: JSON.stringify({
          members: [partnerUsername],
        }),
      });

      if (!res.ok) throw new Error("ルーム作成に失敗");

      const data = await res.json();
      navigate(`/chat/${data.room_id}`);
    } catch (err) {
      console.error("ルーム取得エラー:", err);
    }
  };

  return (
    <div style={{ padding: "1rem" }}>
      <h2>ユーザー一覧</h2>
      <ul style={{ listStyle: "none", padding: 0 }}>
        {users.map((user) => (
          <li
            key={user.id}
            onClick={() => handleClick(user.username)}
            style={{
              cursor: "pointer",
              padding: "8px 12px",
              borderBottom: "1px solid #ccc",
              backgroundColor: "#f9f9f9",
              marginBottom: "4px",
              borderRadius: "4px",
            }}
          >
            {user.username}
          </li>
        ))}
      </ul>
    </div>
  );
}

export default UserList;
