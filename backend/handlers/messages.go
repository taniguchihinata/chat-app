package handlers

import (
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

type Message struct {
	ID        int       `json:"id"`
	FromID    int       `json:"from_id"`
	ToID      int       `json:"to_id"`
	FromName  string    `json:"from_name"`
	ToName    string    `json:"to_name"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

type SendMessageRequest struct {
	ToUsername string `json:"to"`
	Content    string `json:"content"`
}

func getUsernameFromHeader(r *http.Request) (string, error) {
	encoded := r.Header.Get("X-User")
	decodedBytes, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", err
	}
	return string(decodedBytes), nil
}

func MessagesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "GETメソッドのみ許可されています", http.StatusMethodNotAllowed)
		return
	}

	otherUsername := r.URL.Query().Get("with")
	if otherUsername == "" {
		http.Error(w, "クエリ 'with' が必要です", http.StatusBadRequest)
		return
	}

	currentUsername, err := getUsernameFromHeader(r)
	if err != nil {
		log.Println("Base64デコード失敗:", err)
		http.Error(w, "ヘッダーのユーザー名のデコードに失敗しました", http.StatusBadRequest)
		return
	}

	rows, err := DB.Query(r.Context(), `
		SELECT m.id, m.from_id, m.to_id, u1.username AS from_name, u2.username AS to_name, m.content, m.created_at
		FROM messages m
		JOIN users u1 ON m.from_id = u1.id
		JOIN users u2 ON m.to_id = u2.id
		WHERE (u1.username = $1 AND u2.username = $2) OR (u1.username = $2 AND u2.username = $1)
		ORDER BY m.created_at ASC
	`, currentUsername, otherUsername)
	if err != nil {
		log.Println("SQL失敗:", err)
		http.Error(w, "メッセージ取得に失敗しました", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var msg Message
		if err := rows.Scan(&msg.ID, &msg.FromID, &msg.ToID, &msg.FromName, &msg.ToName, &msg.Content, &msg.Timestamp); err != nil {
			log.Println("rows.Scan失敗:", err)
			http.Error(w, "データ読み取り失敗", http.StatusInternalServerError)
			return
		}
		messages = append(messages, msg)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(messages)
}

func SendMessageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POSTメソッドのみ許可されています", http.StatusMethodNotAllowed)
		return
	}

	currentUsername, err := getUsernameFromHeader(r)
	if err != nil {
		http.Error(w, "ヘッダーのユーザー名のデコードに失敗しました", http.StatusBadRequest)
		return
	}

	var req SendMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "リクエストの解析に失敗しました", http.StatusBadRequest)
		return
	}

	var fromID, toID int
	if err := DB.QueryRow(r.Context(), `SELECT id FROM users WHERE username = $1`, currentUsername).Scan(&fromID); err != nil {
		http.Error(w, "送信者取得失敗", http.StatusInternalServerError)
		return
	}
	if err := DB.QueryRow(r.Context(), `SELECT id FROM users WHERE username = $1`, req.ToUsername).Scan(&toID); err != nil {
		http.Error(w, "受信者取得失敗", http.StatusInternalServerError)
		return
	}

	_, err = DB.Exec(r.Context(), `INSERT INTO messages (from_id, to_id, content, created_at) VALUES ($1, $2, $3, NOW())`, fromID, toID, req.Content)
	if err != nil {
		http.Error(w, "メッセージ送信失敗", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("送信成功"))
}
