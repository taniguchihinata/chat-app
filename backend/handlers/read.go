// handlers/read.go
package handlers

import (
	"backend/utils"
	"context"
	"encoding/json"
	"net/http"

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

		w.WriteHeader(http.StatusOK)
	}
}
