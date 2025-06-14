package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"sort"
	"strconv"

	"backend/utils"

	"github.com/jackc/pgx/v5/pgxpool"
)

type RoomRequest struct {
	Members []string `json:"members"`  // ユーザー名の配列
	IsGroup bool     `json:"is_group"` // グループフラグ
	Name    string   `json:"name"`
}

type RoomResponse struct {
	RoomID int `json:"room_id"`
}

// POST /rooms
func GetOrCreateRoomHandler(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username, err := utils.ParseJWTFromRequest(r)
		if err != nil {
			http.Error(w, "認証エラー", http.StatusUnauthorized)
			return
		}

		var req RoomRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "不正なリクエスト形式", http.StatusBadRequest)
			return
		}

		allMembers := append([]string{username}, req.Members...)

		userIDs := []int{}
		for _, uname := range allMembers {
			var uid int
			err := db.QueryRow(context.Background(),
				`SELECT id FROM users WHERE username = $1`, uname).Scan(&uid)
			if err != nil {
				http.Error(w, "ユーザーID取得失敗: "+uname, http.StatusBadRequest)
				return
			}
			userIDs = append(userIDs, uid)
		}

		memberCount := len(userIDs)
		var roomID int

		if !req.IsGroup && memberCount == 2 {
			sort.Ints(userIDs)
			err := db.QueryRow(context.Background(), `
				SELECT rm.room_id
				FROM room_members rm
				JOIN chat_rooms cr ON cr.id = rm.room_id
				WHERE cr.is_group = false
				GROUP BY rm.room_id
				HAVING COUNT(*) = 2 AND
				       ARRAY_AGG(rm.user_id ORDER BY rm.user_id) = ARRAY[$1, $2]::int[]
			`, userIDs[0], userIDs[1]).Scan(&roomID)

			if err == nil {
				json.NewEncoder(w).Encode(RoomResponse{RoomID: roomID})
				return
			}
		}

		if req.IsGroup && memberCount >= 3 && req.Name != "" {
			var exists bool
			err := db.QueryRow(context.Background(), `
				SELECT EXISTS (
					SELECT 1 FROM chat_rooms
					WHERE is_group = true AND room_name = $1
				)
			`, req.Name).Scan(&exists)
			if err == nil && exists {
				http.Error(w, "同名のグループが既に存在します", http.StatusConflict)
				return
			}
		}

		err = db.QueryRow(context.Background(),
			`INSERT INTO chat_rooms (is_group, room_name) VALUES ($1, $2) RETURNING id`,
			req.IsGroup, req.Name).Scan(&roomID)
		if err != nil {
			http.Error(w, "ルーム作成に失敗", http.StatusInternalServerError)
			return
		}

		for _, uid := range userIDs {
			_, err := db.Exec(context.Background(),
				`INSERT INTO room_members (room_id, user_id) VALUES ($1, $2)`,
				roomID, uid)
			if err != nil {
				log.Printf("メンバー登録失敗: room=%d user=%d", roomID, uid)
			}
		}

		json.NewEncoder(w).Encode(RoomResponse{RoomID: roomID})
	}
}

// GET グループ一覧取得
func GetGroupRoomsHandler(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username, err := utils.ParseJWTFromRequest(r)
		if err != nil {
			http.Error(w, "認証エラー", http.StatusUnauthorized)
			return
		}

		var userID int
		err = db.QueryRow(context.Background(),
			`SELECT id FROM users WHERE username = $1`, username).Scan(&userID)
		if err != nil {
			http.Error(w, "ユーザー情報取得失敗", http.StatusInternalServerError)
			return
		}

		rows, err := db.Query(context.Background(), `
			SELECT cr.id, cr.room_name
			FROM chat_rooms cr
			JOIN room_members rm ON cr.id = rm.room_id
			WHERE cr.is_group = true AND rm.user_id = $1
			ORDER BY cr.created_at ASC
		`, userID)
		if err != nil {
			http.Error(w, "DB取得失敗", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		type GroupInfo struct {
			RoomID   int    `json:"room_id"`
			RoomName string `json:"room_name"`
		}

		var groups []GroupInfo
		for rows.Next() {
			var g GroupInfo
			if err := rows.Scan(&g.RoomID, &g.RoomName); err == nil {
				groups = append(groups, g)
			}
		}

		json.NewEncoder(w).Encode(groups)
	}
}

// GETとPOSTを分けて処理する統合ハンドラー
func RoomsHandler(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			GetOrCreateRoomHandler(db)(w, r)
			return
		}
		if r.Method == http.MethodGet && r.URL.Query().Get("type") == "group" {
			GetGroupRoomsHandler(db)(w, r)
			return
		}

		http.Error(w, "不正なリクエスト形式", http.StatusBadRequest)
	}
}

// GET /rooms/{id}
func GetRoomDetailHandler(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, err := utils.ParseJWTFromRequest(r)
		if err != nil {
			http.Error(w, "認証エラー", http.StatusUnauthorized)
			return
		}

		roomIDStr := r.URL.Path[len("/rooms/"):]
		roomID, err := strconv.Atoi(roomIDStr)
		if err != nil {
			http.Error(w, "ルームIDが無効です", http.StatusBadRequest)
			return
		}

		var isGroup bool
		var roomName sql.NullString
		err = db.QueryRow(context.Background(), `
			SELECT is_group, room_name FROM chat_rooms WHERE id = $1
		`, roomID).Scan(&isGroup, &roomName)
		if err != nil {
			http.Error(w, "ルームが見つかりません", http.StatusNotFound)
			return
		}

		rows, err := db.Query(context.Background(), `
			SELECT u.username FROM users u
			JOIN room_members rm ON u.id = rm.user_id
			WHERE rm.room_id = $1
		`, roomID)
		if err != nil {
			http.Error(w, "メンバー取得失敗", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var members []string
		for rows.Next() {
			var uname string
			if err := rows.Scan(&uname); err == nil {
				members = append(members, uname)
			}
		}

		result := map[string]interface{}{
			"type":    "direct",
			"users":   members,
			"name":    "",
			"room_id": roomID,
		}

		if isGroup {
			result["type"] = "group"
			result["name"] = roomName.String
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	}
}
