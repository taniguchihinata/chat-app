// ユーザー取得のバックエンド処理
package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
)

// 関数定義
func UsersHandler(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "GETメソッドのみ許可されています", http.StatusMethodNotAllowed)
			return
		}

		//データベースからユーザー一覧を取得
		rows, err := db.Query(r.Context(), `SELECT id, username FROM users`)
		if err != nil {
			http.Error(w, "ユーザー一覧の取得に失敗しました", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		//データをスライスに格納
		//スライスは柔軟な配列のようなデータ構造、長さが可変
		var users []map[string]interface{}
		for rows.Next() {
			var id int
			var username string
			if err := rows.Scan(&id, &username); err != nil {
				http.Error(w, "データの読み取りに失敗しました", http.StatusInternalServerError)
				return
			}
			users = append(users, map[string]interface{}{
				"id":       id,
				"username": username,
			})
		}

		//レスポンスの送信
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(users)
	}
}
