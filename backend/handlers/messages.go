// handlers/messages.go
package handlers

import (
	"backend/utils"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// メッセージ取得用構造体
type Message struct {
	ID        int    `json:"id"`
	SenderID  int    `json:"sender_id"`
	Username  string `json:"username"`
	Text      string `json:"text"`
	CreatedAt string `json:"created_at"`
	Image     string `json:"image,omitempty"`
}

// 送信リクエスト用構造体
type SendMessageRequest struct {
	RoomID int    `json:"room_id"`
	Text   string `json:"text"`
}

// メッセージの取得ハンドラー
func GetMessagesHandler(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		//認証チェック
		username, err := utils.ParseJWTFromRequest(r)
		if err != nil {
			log.Println("JWT検証失敗:", err)
			http.Error(w, "認証エラー", http.StatusUnauthorized)
			return
		}
		log.Println("メッセージ取得ユーザー:", username)

		//クエリパラメータの取得と変換
		roomIDStr := r.URL.Query().Get("room")
		roomID, err := strconv.Atoi(roomIDStr)
		if err != nil {
			http.Error(w, "roomパラメータが無効です", http.StatusBadRequest)
			return
		}

		//データベースクエリと整形
		rows, err := db.Query(
			context.Background(),
			`SELECT m.id, m.sender_id, u.username, m.text,COALESCE(a.file_url, '') as image, m.created_at
			 FROM messages m
			 JOIN users u ON m.sender_id = u.id
			 LEFT JOIN message_attachments a ON m.id = a.message_id
			 WHERE m.room_id = $1
			 ORDER BY m.created_at ASC`, roomID,
		)
		if err != nil {
			log.Println("メッセージ取得失敗:", err)
			http.Error(w, "メッセージ取得に失敗しました", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		//メッセージを取得
		var messages []Message
		for rows.Next() {
			var msg Message
			var createdAt time.Time
			err := rows.Scan(&msg.ID, &msg.SenderID, &msg.Username, &msg.Text, &msg.Image, &createdAt)
			if err != nil {
				log.Println("行スキャン失敗:", err)
				continue
			}
			//レスポンス整形と返却
			msg.CreatedAt = createdAt.Format(time.RFC3339)
			messages = append(messages, msg)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(messages)
	}
}

// メッセージの送信ハンドラー
func SendMessageHandler(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		//JWT認証
		username, err := utils.ParseJWTFromRequest(r)
		if err != nil {
			log.Println("JWT検証失敗:", err)
			http.Error(w, "認証エラー", http.StatusUnauthorized)
			return
		}
		log.Println("送信者（JWT）:", username)

		//リクエストボディの読み取り
		var req SendMessageRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Println("リクエストデコード失敗:", err)
			http.Error(w, "無効なリクエスト", http.StatusBadRequest)
			return
		}

		//ユーザーIDの取得
		var senderID int
		err = db.QueryRow(context.Background(), "SELECT id FROM users WHERE username=$1", username).Scan(&senderID)
		if err != nil {
			log.Println("ユーザーID取得失敗:", err)
			http.Error(w, "ユーザーが見つかりません", http.StatusUnauthorized)
			return
		}

		_, err = db.Exec(
			context.Background(),
			//メッセージ挿入
			"INSERT INTO messages (room_id, sender_id, text, created_at) VALUES ($1, $2, $3, $4)",
			req.RoomID, senderID, req.Text, time.Now(),
		)
		if err != nil {
			log.Println("メッセージ挿入失敗:", err)
			http.Error(w, "送信失敗", http.StatusInternalServerError)
			return
		}

		//レスポンスを返す
		w.WriteHeader(http.StatusCreated)
	}
}
