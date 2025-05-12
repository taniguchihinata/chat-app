package handlers

import (
	"encoding/json"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POSTメソッドのみ許可されています", http.StatusMethodNotAllowed)
		return
	}

	var req SignupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "リクエストの形式が不正です", http.StatusBadRequest)
		return
	}

	var passwordHash string
	err := DB.QueryRow(r.Context(), `SELECT password_hash FROM users WHERE username = $1`, req.Username).Scan(&passwordHash)
	if err != nil || bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password)) != nil {
		http.Error(w, "ユーザー名またはパスワードが不正です", http.StatusUnauthorized)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ログイン成功"))
}
