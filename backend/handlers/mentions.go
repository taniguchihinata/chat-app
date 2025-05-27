package handlers

import (
	"backend/utils"
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type MentionResponse struct {
	MessageID  int    `json:"message_id"`
	RoomID     int    `json:"room_id"`
	Text       string `json:"text"`
	CreatedAt  string `json:"created_at"`
	SenderName string `json:"sender_name"`
	IsRead     bool   `json:"is_read"`
}

func GetMentionsHandler(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username, err := utils.ParseJWTFromRequest(r)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		var userID int
		err = db.QueryRow(context.Background(),
			"SELECT id FROM users WHERE username = $1", username).Scan(&userID)
		if err != nil {
			http.Error(w, "User not found", http.StatusInternalServerError)
			return
		}

		rows, err := db.Query(context.Background(), `
  			SELECT m.message_id, msg.room_id, msg.text, msg.created_at, u.username,
         		   CASE WHEN r.read_at IS NULL THEN false ELSE true END AS is_read
  			FROM mentions m
  			JOIN messages msg ON m.message_id = msg.id
  			JOIN users u ON msg.sender_id = u.id
  			LEFT JOIN message_reads r ON r.message_id = m.message_id AND r.user_id = $1
  			WHERE m.mention_target_id = $1
  			ORDER BY msg.created_at DESC
		`, userID)
		if err != nil {
			http.Error(w, "Failed to get mentions", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var mentions []MentionResponse
		for rows.Next() {
			var m MentionResponse
			var created time.Time
			var isRead bool
			if err := rows.Scan(&m.MessageID, &m.RoomID, &m.Text, &created, &m.SenderName, &m.IsRead); err == nil {
				m.CreatedAt = created.Format(time.RFC3339)
				m.IsRead = isRead
				mentions = append(mentions, m)
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mentions)
	}
}

func MarkMentionAsReadHandler(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username, err := utils.ParseJWTFromRequest(r)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		var userID int
		err = db.QueryRow(context.Background(),
			"SELECT id FROM users WHERE username=$1", username).Scan(&userID)
		if err != nil {
			http.Error(w, "ユーザー取得失敗", http.StatusInternalServerError)
			return
		}

		var req struct {
			MessageID int `json:"message_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "無効なリクエスト", http.StatusBadRequest)
			return
		}

		_, err = db.Exec(context.Background(),
			"UPDATE mentions SET is_read = TRUE WHERE mention_target_id=$1 AND message_id=$2",
			userID, req.MessageID)
		if err != nil {
			http.Error(w, "更新失敗", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
