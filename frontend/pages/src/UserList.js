import React, { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";

function UserList({ reloadFlag}) {
  const [users, setUsers] = useState([]);
  const [groups, setGroups] = useState([]);
  const [unreadCounts, setUnreadCounts] = useState({});
  const [privateRoomMap, setPrivateRoomMap] = useState({});
  const [privateUnreadCounts, setPrivateUnreadCounts] = useState({});
  const navigate = useNavigate();
  
  const myUsername = localStorage.getItem("username");
  
  // ユーザーリスト一覧を取得
  useEffect(() => {
    const fetchUsers = async () => {
      try {
        const res = await fetch("http://localhost:8081/users");
        const data = await res.json();
        const filtered = Array.isArray(data)
          ? data.filter((user) => user.username !== myUsername)
          : [];
        setUsers(filtered);
      } catch (err) {
        console.error("ユーザー取得失敗:", err);
        setUsers([]);
      }
    };

    fetchUsers();
  }, [myUsername]);

  // グループ一覧を取得
  useEffect(() => {
    const fetchGroups = async () => {
      try {
        const token = localStorage.getItem("token");
        const res = await fetch("http://localhost:8081/rooms?type=group", {
          headers: { Authorization: `Bearer ${token}` },
        });
        const data = await res.json();
        setGroups(Array.isArray(data) ? data : []);
      } catch (err) {
        console.error("グループ取得失敗:", err);
        setGroups([]);
      }
    };

    fetchGroups();
  }, []);

  // 未読件数の再取得（reloadFlagトリガー）
  useEffect(() => {
    const token = localStorage.getItem("token");

    const fetchUnreadCounts = async () => {
      const results = await Promise.all(
        groups.map(async (group) => {
          try {
            const res = await fetch(`http://localhost:8081/unread_count?room=${group.room_id}`, {
              headers: { Authorization: `Bearer ${token}` },
            });
            const data = await res.json();
            return { roomId: group.room_id, count: data.count };
          } catch {
            return { roomId: group.room_id, count: 0 };
          }
        })
      );
      const map = {};
      results.forEach(({ roomId, count }) => {
        map[roomId] = count;
      });
      setUnreadCounts(map);
    };

    const fetchPrivateUnread = async () => {
      const roomMap = {};
      const countMap = {};

      for (const user of users) {
        try {
          const res = await fetch("http://localhost:8081/rooms", {
            method: "POST",
            headers: {
              Authorization: `Bearer ${token}`,
              "Content-Type": "application/json",
            },
            body: JSON.stringify({ members: [user.username] }),
          });

          if (!res.ok) continue;

          const data = await res.json();
          const roomId = data.room_id;
          roomMap[user.username] = roomId;

          const countRes = await fetch(`http://localhost:8081/unread_count?room=${roomId}`, {
            headers: { Authorization: `Bearer ${token}` },
          });

          const countData = await countRes.json();
          countMap[user.username] = countData.count;
        } catch {}
      }

      setPrivateRoomMap(roomMap);
      setPrivateUnreadCounts(countMap);
    };

    if (groups.length > 0) fetchUnreadCounts();
    if (users.length > 0) fetchPrivateUnread();
  }, [reloadFlag, groups, users]);

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
      <h2 className="user-list-title">ユーザー一覧</h2>
      <ul style ={{ listStyle: "none", padding: 0 }}>
        {Array.isArray(users) &&
          users.map((user) => (
            <li
              key={user.id}
              onClick={() => handleClick(user.username)}
              className="user-name"
              style={{ display: "flex", justifyContent: "space-between", alignItems: "center" }}
            >
              <span>{user.username}</span>
              {privateUnreadCounts[user.username] > 0 && (
                <span className="badge">{privateUnreadCounts[user.username]}</span>
              )}
            </li>
          ))}
      </ul>
      <h2 className="user-list-title" style={{ alignItems: "center", justifyContent: "space-between" }}>
        グループ一覧
        <button
          onClick={() => navigate("/groups/create")}
          style={{ fontSize: "0.8rem", padding: "4px 8px", marginLeft: "12px", fontWeight: 500, left: "500px"}}
        >
        新規グループ作成
        </button>
      </h2>
      <ul style={{ listStyle: "none", padding: 0 }}>
        {Array.isArray(groups) &&
          groups.map((group) => (
            <li
              key={group.room_id}
              onClick={() => navigate(`/chat/${group.room_id}`)}
              className="user-name"
              style={{ display: "flex", justifyContent: "space-between", alignItems: "center" }}
            >
              <span>{group.room_name || "(名前なしのグループ)"}</span>
              {unreadCounts[group.room_id] > 0 && (
                <span className="badge">{unreadCounts[group.room_id]}</span>
              )}
            </li>

          ))
        }
      </ul>
    </div>
  );
}

export default UserList;
