// サインアップシステムバックエンド
package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type SignupRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// 引数がデータベース接続、外部で接続してこの関数に渡すので再利用性が高い
func SignupHandler(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		//POST以外は弾く
		if r.Method != http.MethodPost {
			http.Error(w, "メソッドはPOSTのみ許可されています", http.StatusMethodNotAllowed)
			return
		}

		//リクエストボディのデコード
		var req SignupRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "リクエストの形式が不正です", http.StatusBadRequest)
			return
		}

		//パスワードのハッシュ化
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, "パスワードのハッシュ化に失敗しました", http.StatusInternalServerError)
			return
		}

		//ユーザー情報をデータベースに登録
		_, err = db.Exec(r.Context(), `INSERT INTO users (username, password_hash) VALUES ($1, $2)`, req.Username, string(hashedPassword))
		if err != nil {
			log.Println("INSERT失敗:", err)
			http.Error(w, "ユーザー登録に失敗しました", http.StatusInternalServerError)
			return
		}

		//成功したらレスポンスを返す
		log.Printf("登録ユーザー: %s", req.Username)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ユーザー登録成功"))
	}
}
