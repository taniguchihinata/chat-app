import React, { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";

function UserList() {
  const [users, setUsers] = useState([]);
  const navigate = useNavigate();

  useEffect(() => {
    fetch("http://localhost:8081/users")
      .then((res) => res.json())
      .then((data) => {
        const currentUser = localStorage.getItem("username");
        const filtered = data.filter(
          (user) => user.username !== currentUser
        );
        setUsers(filtered);
      });
  }, []);

  const handleClick = async (partnerUsername) => {
    const currentUser = localStorage.getItem("username");
    const encodedUser = btoa(unescape(encodeURIComponent(currentUser)));

    try {
      const res = await fetch("http://localhost:8081/rooms", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          "X-User": encodedUser,
        },
        body: JSON.stringify({ partner: partnerUsername }),
      });

      if (!res.ok) throw new Error("ルーム作成に失敗");

      const data = await res.json();
      navigate(`/chat/${data.room_id}`);
    } catch (err) {
      console.error("ルーム取得エラー:", err);
    }
  };

  const handleLogout = () => {
    localStorage.removeItem("username");
    navigate("/");
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
      <button onClick={handleLogout} style={{ marginBottom: "1rem" }}>
        ログアウト
      </button>
    </div>
  );
}

export default UserList;
