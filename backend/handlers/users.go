package handlers

import (
	"encoding/json"
	"net/http"
)

func UsersHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "GETメソッドのみ許可されています", http.StatusMethodNotAllowed)
		return
	}

	rows, err := DB.Query(r.Context(), `SELECT id, username FROM users`)
	if err != nil {
		http.Error(w, "ユーザー一覧の取得に失敗しました", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}
