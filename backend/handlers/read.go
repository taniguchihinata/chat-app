// handlers/read.go
package handlers

import (
	"backend/utils"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/jackc/pgx/v5/pgxpool"
)

func MarkAsReadHandler(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// JWT から username を取得
		username, err := utils.ParseJWTFromRequest(r)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// リクエストボディから message_id を取得
		var req struct {
			MessageID int `json:"message_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		// ユーザーIDを取得
		var userID int
		err = db.QueryRow(context.Background(),
			"SELECT id FROM users WHERE username = $1", username,
		).Scan(&userID)
		if err != nil {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}

		// message_reads に登録（すでに既読なら無視）
		_, err = db.Exec(context.Background(),
			`INSERT INTO message_reads (message_id, reader_id)
			 VALUES ($1, $2)
			 ON CONFLICT DO NOTHING`,
			req.MessageID, userID,
		)
		if err != nil {
			http.Error(w, "Failed to insert", http.StatusInternalServerError)
			return
		}

		// メッセージのルームIDを取得
		var roomID int
		err = db.QueryRow(context.Background(), `SELECT room_id FROM messages WHERE id = $1`, req.MessageID).Scan(&roomID)
		if err != nil {
			log.Println("ルームID取得失敗:", err)
			w.WriteHeader(http.StatusOK)
			return
		}

		// WebSocketで通知
		RoomClientsMutex.Lock()
		for _, c := range RoomClients[roomID] {
			_ = c.Conn.WriteJSON(map[string]interface{}{
				"type":       "read",
				"room_id":    roomID,
				"username":   username,
				"message_id": req.MessageID,
			})
		}
		RoomClientsMutex.Unlock()

		w.WriteHeader(http.StatusOK)
	}
}

// 既読状態取得
func GetReadStatusHandler(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username, err := utils.ParseJWTFromRequest(r)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		roomIDStr := r.URL.Query().Get("room")
		roomID, err := strconv.Atoi(roomIDStr)
		if err != nil {
			http.Error(w, "Invalid room ID", http.StatusBadRequest)
			return
		}

		// 自分が送信したメッセージIDのうち、既読がついているもの
		rows, err := db.Query(context.Background(), `
			SELECT mr.message_id
			FROM message_reads mr
			JOIN messages m ON mr.message_id = m.id
			JOIN users u ON u.username = $1
			WHERE m.sender_id = u.id AND m.room_id = $2
		`, username, roomID)
		if err != nil {
			log.Printf("message_reads挿入失敗: %v", err)
			http.Error(w, "Failed to get read status", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var ids []int
		for rows.Next() {
			var id int
			if err := rows.Scan(&id); err == nil {
				ids = append(ids, id)
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ids)
	}
}
