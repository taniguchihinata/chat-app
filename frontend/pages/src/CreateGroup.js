import React, { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";

function CreateGroup() {
  const [users, setUsers] = useState([]);
  const [selectedUsers, setSelectedUsers] = useState([]);
  const [groupName, setGroupName] = useState(""); // ← グループ名
  const navigate = useNavigate();

  const myUsername = localStorage.getItem("username");

  // 自分以外のユーザー一覧取得
  useEffect(() => {
    const fetchUsers = async () => {
      const res = await fetch("http://localhost:8081/users");
      const data = await res.json();
      const filtered = data.filter((u) => u.username !== myUsername);
      setUsers(filtered);
    };

    fetchUsers();
  }, [myUsername]);

  const handleToggle = (username) => {
    if (selectedUsers.includes(username)) {
      setSelectedUsers(selectedUsers.filter((u) => u !== username));
    } else {
      setSelectedUsers([...selectedUsers, username]);
    }
  };

  const handleCreateGroup = async () => {
    const token = localStorage.getItem("token");

    try {
      const res = await fetch("http://localhost:8081/rooms", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          "Authorization": `Bearer ${token}`,
        },
        body: JSON.stringify({
          members: selectedUsers,
          is_group: true,
          name: groupName.trim(),
        }),
      });

      if (res.status === 409) {
        alert("同名のグループが既に存在します");
        return;
      }

      if (!res.ok) throw new Error("グループ作成に失敗");

      const data = await res.json();
      navigate(`/chat/${data.room_id}`);
    } catch (err) {
      console.error("グループ作成エラー:", err);
      alert("グループ作成に失敗しました");
    }
  };

  return (
    <div style={{ padding: "1rem" }}>
      <h2>グループチャット作成</h2>

      {/* グループ名の入力欄 */}
      <div style={{ marginBottom: "1rem" }}>
        <label>
          グループ名：
          <input
            type="text"
            value={groupName}
            onChange={(e) => setGroupName(e.target.value)}
            style={{ marginLeft: "8px", padding: "4px" }}
          />
        </label>
      </div>

      <ul style={{ listStyle: "none", padding: 0 }}>
        {users.map((user) => (
          <li key={user.id} style={{ marginBottom: "8px" }}>
            <label>
              <input
                type="checkbox"
                checked={selectedUsers.includes(user.username)}
                onChange={() => handleToggle(user.username)}
              />
              {user.username}
            </label>
          </li>
        ))}
      </ul>

      <button
        onClick={handleCreateGroup}
        disabled={selectedUsers.length === 0 || groupName.trim() === ""}
        style={{ marginTop: "10px" }}
      >
        グループを作成
      </button>
    </div>
  );
}

export default CreateGroup;
