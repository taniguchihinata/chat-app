// handlers/messages.go
package handlers

import (
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// メッセージ取得用構造体
type Message struct {
	ID        int       `json:"id"`
	Sender    string    `json:"from_name"`
	Text      string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

// 送信リクエスト用構造体
type SendMessageRequest struct {
	RoomID int    `json:"room_id"`
	Text   string `json:"text"`
}

func decodeUsername(r *http.Request) (string, error) {
	encoded := r.Header.Get("X-User")
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", err
	}
	return string(decoded), nil
}

// メッセージの取得関数
func GetMessagesHandler(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		roomIDStr := r.URL.Query().Get("room")
		roomID, err := strconv.Atoi(roomIDStr)
		if err != nil {
			http.Error(w, "invalid room id", http.StatusBadRequest)
			return
		}

		rows, err := db.Query(r.Context(), `
			SELECT m.id, m.sender_id, u.username, m.text, m.created_at
			FROM messages m
			JOIN users u ON m.sender_id = u.id
			WHERE m.room_id = $1
			ORDER BY m.created_at ASC
		`, roomID)
		if err != nil {
			log.Println("メッセージ取得失敗:", err)
			http.Error(w, "メッセージ取得に失敗", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var messages []Message
		for rows.Next() {
			var m Message
			var ignoredSenderID int
			if err := rows.Scan(&m.ID, &ignoredSenderID, &m.Sender, &m.Text, &m.CreatedAt); err != nil {
				log.Println("行スキャン失敗:", err)
				continue
			}
			messages = append(messages, m)
		}
		json.NewEncoder(w).Encode(messages)
	}
}

// メッセージの送信関数
func SendMessageHandler(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username, err := decodeUsername(r)
		if err != nil {
			http.Error(w, "ユーザー認証失敗", http.StatusUnauthorized)
			return
		}

		var req SendMessageRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "リクエスト不正", http.StatusBadRequest)
			return
		}

		var userID int
		err = db.QueryRow(r.Context(), `SELECT id FROM users WHERE username = $1`, username).Scan(&userID)
		if err != nil {
			log.Println("ユーザーID取得失敗:", err)
			http.Error(w, "ユーザー取得に失敗", http.StatusInternalServerError)
			return
		}

		_, err = db.Exec(r.Context(), `
			INSERT INTO messages (sender_id, room_id, text)
			VALUES ($1, $2, $3)
		`, userID, req.RoomID, req.Text)
		if err != nil {
			log.Println("メッセージ挿入失敗:", err)
			http.Error(w, "メッセージ送信に失敗", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
	}
}
