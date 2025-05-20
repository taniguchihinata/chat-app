package handlers

import (
	"backend/utils"
	"context"
	"encoding/json"
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RoomRequest struct {
	Members []string `json:"members"` // 自分を除いた他のユーザー名
}

type RoomResponse struct {
	RoomID int `json:"room_id"`
}

func GetOrCreateRoomHandler(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username, err := utils.ParseJWTFromRequest(r)
		if err != nil {
			http.Error(w, "認証エラー", http.StatusUnauthorized)
			return
		}

		var req RoomRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "リクエスト形式エラー", http.StatusBadRequest)
			return
		}

		// 全メンバー（自分＋相手）
		allUsernames := append([]string{username}, req.Members...)
		isGroup := len(allUsernames) > 2

		// ユーザーID配列を作成
		userIDs := make([]int32, 0, len(allUsernames)) // pgtype.Int4Array は int32

		for _, uname := range allUsernames {
			var uid int32
			err := db.QueryRow(context.Background(), "SELECT id FROM users WHERE username=$1", uname).Scan(&uid)
			if err != nil {
				http.Error(w, "ユーザーが見つかりません: "+uname, http.StatusNotFound)
				return
			}
			userIDs = append(userIDs, uid)
		}

		var roomID int

		if !isGroup {
			// 1対1ルームの重複チェック
			err = db.QueryRow(context.Background(), `
				SELECT rm.room_id
				FROM room_members rm
				JOIN chat_rooms cr ON cr.id = rm.room_id
				WHERE rm.user_id = ANY($1) AND cr.is_group = false
				GROUP BY rm.room_id
				HAVING COUNT(DISTINCT rm.user_id) = 2
			`, userIDs).Scan(&roomID)

			if err == nil {
				// 既存ルームがあれば返す
				json.NewEncoder(w).Encode(RoomResponse{RoomID: roomID})
				return
			}
		}

		// 新しいルームを作成
		err = db.QueryRow(context.Background(),
			`INSERT INTO chat_rooms (is_group, created_at, updated_at)
			 VALUES ($1, now(), now()) RETURNING id`,
			isGroup,
		).Scan(&roomID)
		if err != nil {
			http.Error(w, "ルーム作成失敗", http.StatusInternalServerError)
			return
		}

		// メンバー登録
		batch := &pgx.Batch{}
		for _, uid := range userIDs {
			batch.Queue(`INSERT INTO room_members (room_id, user_id, joined_at) VALUES ($1, $2, now())`, roomID, uid)
		}
		br := db.SendBatch(context.Background(), batch)
		defer br.Close()

		if err := br.Close(); err != nil {
			http.Error(w, "メンバー登録失敗", http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(RoomResponse{RoomID: roomID})
	}
}
