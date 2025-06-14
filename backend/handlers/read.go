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
			`INSERT INTO message_reads (message_id, user_id)
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

// 全メッセージの既読ユーザーを取得
func GetFullReadStatusHandler(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		roomIDStr := r.URL.Query().Get("room")
		roomID, err := strconv.Atoi(roomIDStr)
		if err != nil {
			http.Error(w, "Invalid room ID", http.StatusBadRequest)
			return
		}

		rows, err := db.Query(context.Background(), `
			SELECT mr.message_id, u.username
			FROM message_reads mr
			JOIN users u ON mr.user_id = u.id
			JOIN messages m ON mr.message_id = m.id
			WHERE m.room_id = $1
		`, roomID)
		if err != nil {
			log.Printf("read_status_full取得失敗: %v", err)
			http.Error(w, "Failed to fetch", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		result := make(map[int][]string)
		for rows.Next() {
			var msgID int
			var username string
			if err := rows.Scan(&msgID, &username); err == nil {
				result[msgID] = append(result[msgID], username)
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	}
}

func GetUnreadCountHandler(db *pgxpool.Pool) http.HandlerFunc {
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

		var userID int
		err = db.QueryRow(context.Background(),
			"SELECT id FROM users WHERE username = $1", username,
		).Scan(&userID)
		if err != nil {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}

		// 未読件数の集計クエリ
		var count int
		err = db.QueryRow(context.Background(), `
			SELECT COUNT(*)
			FROM messages m
			WHERE m.room_id = $1
			  AND m.sender_id != $2
			  AND NOT EXISTS (
			    SELECT 1 FROM message_reads mr
			    WHERE mr.message_id = m.id AND mr.user_id = $2
			  )
		`, roomID, userID).Scan(&count)
		if err != nil {
			http.Error(w, "Failed to count unread", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]int{"count": count})
	}
}

func MarkAllAsReadHandler(db *pgxpool.Pool) http.HandlerFunc {
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

		var userID int
		err = db.QueryRow(context.Background(),
			"SELECT id FROM users WHERE username = $1", username,
		).Scan(&userID)
		if err != nil {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}

		// まだ既読になっていないメッセージを一括登録
		_, err = db.Exec(context.Background(), `
      INSERT INTO message_reads (message_id, user_id)
      SELECT m.id, $1
      FROM messages m
      WHERE m.room_id = $2
        AND m.sender_id != $1
        AND NOT EXISTS (
          SELECT 1 FROM message_reads mr
          WHERE mr.message_id = m.id AND mr.user_id = $1
        )
    `, userID, roomID)

		if err != nil {
			http.Error(w, "Failed to mark all as read", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
