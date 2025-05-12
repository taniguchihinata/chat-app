package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

var DB *pgxpool.Pool

type SignupRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func SignupHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "メソッドはPOSTのみ許可されています", http.StatusMethodNotAllowed)
		return
	}

	var req SignupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "リクエストの形式が不正です", http.StatusBadRequest)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "パスワードのハッシュ化に失敗しました", http.StatusInternalServerError)
		return
	}

	_, err = DB.Exec(r.Context(), `INSERT INTO users (username, password_hash) VALUES ($1, $2)`, req.Username, string(hashedPassword))
	if err != nil {
		log.Println("INSERT失敗:", err)
		http.Error(w, "ユーザー登録に失敗しました", http.StatusInternalServerError)
		return
	}

	log.Printf("登録ユーザー: %s", req.Username)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ユーザー登録成功"))
}
