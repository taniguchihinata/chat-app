// チャットルームのデータベースを作成
package handlers

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
)

type RoomRequest struct {
	PartnerUsername string `json:"partner"`
}

type RoomResponse struct {
	RoomID int `json:"room_id"`
}

func GetOrCreateRoomHandler(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "POSTメソッドのみ許可されています", http.StatusMethodNotAllowed)
			return
		}

		// ヘッダーから自分のusernameを取得
		encodedUser := r.Header.Get("X-User")
		decodedBytes, err := base64.StdEncoding.DecodeString(encodedUser)
		if err != nil {
			http.Error(w, "ユーザー名デコード失敗", http.StatusBadRequest)
			return
		}
		username := string(decodedBytes)

		// ボディから相手のusernameを取得
		var req RoomRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "リクエストが不正です", http.StatusBadRequest)
			return
		}

		// 送信者と受信者のIDを取得
		var myID, partnerID int
		if err := db.QueryRow(context.Background(), `SELECT id FROM users WHERE username = $1`, username).Scan(&myID); err != nil {
			http.Error(w, "送信者ユーザーが見つかりません", http.StatusBadRequest)
			return
		}
		if err := db.QueryRow(context.Background(), `SELECT id FROM users WHERE username = $1`, req.PartnerUsername).Scan(&partnerID); err != nil {
			http.Error(w, "相手ユーザーが見つかりません", http.StatusBadRequest)
			return
		}

		// 既存のルームを探す（両ユーザーが参加しているルーム）
		var roomID int
		query := `
			SELECT rm1.room_id
			FROM room_members rm1
			JOIN room_members rm2 ON rm1.room_id = rm2.room_id
			WHERE rm1.user_id = $1 AND rm2.user_id = $2
			LIMIT 1`
		err = db.QueryRow(context.Background(), query, myID, partnerID).Scan(&roomID)
		if err == nil {
			log.Println("既存ルームを使用: room_id =", roomID)
			json.NewEncoder(w).Encode(RoomResponse{RoomID: roomID})
			return
		}

		// なければ新規作成
		tx, err := db.Begin(context.Background())
		if err != nil {
			http.Error(w, "トランザクション開始失敗", http.StatusInternalServerError)
			return
		}
		defer tx.Rollback(context.Background())

		err = tx.QueryRow(context.Background(), `INSERT INTO chat_rooms DEFAULT VALUES RETURNING id`).Scan(&roomID)
		if err != nil {
			http.Error(w, "ルーム作成失敗", http.StatusInternalServerError)
			return
		}
		_, err = tx.Exec(context.Background(), `INSERT INTO room_members (room_id, user_id) VALUES ($1, $2), ($1, $3)`, roomID, myID, partnerID)
		if err != nil {
			http.Error(w, "ルームメンバー追加失敗", http.StatusInternalServerError)
			return
		}
		if err := tx.Commit(context.Background()); err != nil {
			http.Error(w, "コミット失敗", http.StatusInternalServerError)
			return
		}

		log.Println("新規ルーム作成: room_id =", roomID)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(RoomResponse{RoomID: roomID})
	}
}
