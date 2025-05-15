// ログインシステムバックエンド
package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"backend/utils"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func LoginHandler(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var creds Credentials
		if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
			log.Println("JSONデコード失敗:", err)
			http.Error(w, "無効なリクエスト", http.StatusBadRequest)
			return
		}
		log.Println("ログイン試行:", creds.Username)

		var hashedPassword string
		err := db.QueryRow(context.Background(), "SELECT password_hash FROM users WHERE username=$1", creds.Username).Scan(&hashedPassword)
		if err != nil {
			if err == sql.ErrNoRows {
				log.Println("ユーザーが存在しません:", creds.Username)
				http.Error(w, "ユーザーが存在しません", http.StatusUnauthorized)
			} else {
				log.Println("SQLエラー:", err)
				http.Error(w, "サーバーエラー", http.StatusInternalServerError)
			}
			return
		}

		if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(creds.Password)); err != nil {
			log.Println("パスワード不一致:", creds.Username)
			http.Error(w, "パスワードが間違っています", http.StatusUnauthorized)
			return
		}
		log.Println("パスワード認証成功:", creds.Username)

		token, err := utils.GenerateJWT(creds.Username)
		if err != nil {
			log.Printf("トークン生成失敗:%+v\n", err)
			http.Error(w, "トークン生成失敗", http.StatusInternalServerError)
			return
		}
		log.Println("トークン生成成功:", token)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"token": token,
		})
	}
}
