// handlers/messages.go
package handlers

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/jackc/pgx/v5/pgxpool"
)

// メッセージ取得用構造体
type Message struct {
	ID        int    `json:"id"`
	UserID    int    `json:"user_id"`   //送信者ID
	FromName  string `json:"from_name"` //送信者名
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"` //ISO文字列
}

// 送信リクエスト用構造体
type SendMessageRequest struct {
	RoomID  int    `json:"room_id"`
	Content string `json:"content"`
}

// メッセージの取得関数
func GetMessagesHandler(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "GETメソッドのみ許可されています", http.StatusMethodNotAllowed)
			return
		}

		roomIDStr := r.URL.Query().Get("room")
		roomID, err := strconv.Atoi(roomIDStr)
		if err != nil {
			http.Error(w, "room_idが無効です", http.StatusBadRequest)
			return
		}

		rows, err := db.Query(context.Background(), `
			SELECT m.id, m.user_id, u.username, m.content, m.created_at
			FROM messages m
			JOIN users u ON m.user_id = u.id
			WHERE m.room_id = $1
			ORDER BY m.created_at ASC
		`, roomID)
		if err != nil {
			log.Println("メッセージ取得失敗:", err)
			http.Error(w, "メッセージ取得失敗", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var messages []Message
		for rows.Next() {
			var m Message
			if err := rows.Scan(&m.ID, &m.UserID, &m.FromName, &m.Content, &m.CreatedAt); err != nil {
				log.Println("スキャン失敗:", err)
				http.Error(w, "読み取り失敗", http.StatusInternalServerError)
				return
			}
			messages = append(messages, m)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(messages)
	}
}

// メッセージの送信関数
func SendMessageHandler(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "POSTメソッドのみ許可されています", http.StatusMethodNotAllowed)
			return
		}

		encodedUser := r.Header.Get("X-User")
		decodedBytes, err := base64.StdEncoding.DecodeString(encodedUser)
		if err != nil {
			http.Error(w, "ユーザー名デコード失敗", http.StatusBadRequest)
			return
		}
		username := string(decodedBytes)

		var userID int
		if err := db.QueryRow(context.Background(), `SELECT id FROM users WHERE username = $1`, username).Scan(&userID); err != nil {
			http.Error(w, "ユーザー取得失敗", http.StatusBadRequest)
			return
		}

		var req SendMessageRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "JSONパース失敗", http.StatusBadRequest)
			return
		}

		_, err = db.Exec(context.Background(), `
			INSERT INTO messages (user_id, room_id, content, created_at)
			VALUES ($1, $2, $3, NOW())
		`, userID, req.RoomID, req.Content)
		if err != nil {
			log.Println("メッセージ挿入失敗:", err)
			http.Error(w, "メッセージ保存失敗", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("送信成功"))
	}
}
