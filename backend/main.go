// ✅ 修正済み main.go（GET /messages, POST /send, ログ追加）
package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

var db *pgx.Conn

type SignupRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func signupHandler(w http.ResponseWriter, r *http.Request) {
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

	_, err = db.Exec(r.Context(), `
        INSERT INTO users (username, password_hash)
        VALUES ($1, $2)
    `, req.Username, string(hashedPassword))
	if err != nil {
		log.Println("INSERT失敗:", err)
		http.Error(w, "ユーザー登録に失敗しました", http.StatusInternalServerError)
		return
	}

	log.Printf("登録ユーザー: %s", req.Username)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ユーザー登録成功"))
}

func usersHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "GETメソッドのみ許可されています", http.StatusMethodNotAllowed)
		return
	}

	rows, err := db.Query(r.Context(), `SELECT id, username FROM users`)
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

func withCORS(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-User")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		handler.ServeHTTP(w, r)
	})
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
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
	err := db.QueryRow(r.Context(), `
        SELECT password_hash FROM users WHERE username = $1
    `, req.Username).Scan(&passwordHash)
	if err != nil || bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password)) != nil {
		http.Error(w, "ユーザー名またはパスワードが不正です", http.StatusUnauthorized)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ログイン成功"))
}

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

func messagesHandler(w http.ResponseWriter, r *http.Request) {
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

	log.Println("メッセージ取得（自分）:", currentUsername)
	log.Println("メッセージ取得（相手）:", otherUsername)

	rows, err := db.Query(r.Context(), `
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

	log.Printf("取得メッセージ数: %d", len(messages))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(messages)
}

func sendMessageHandler(w http.ResponseWriter, r *http.Request) {
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
	if err := db.QueryRow(r.Context(), `SELECT id FROM users WHERE username = $1`, currentUsername).Scan(&fromID); err != nil {
		http.Error(w, "送信者取得失敗", http.StatusInternalServerError)
		return
	}
	if err := db.QueryRow(r.Context(), `SELECT id FROM users WHERE username = $1`, req.ToUsername).Scan(&toID); err != nil {
		http.Error(w, "受信者取得失敗", http.StatusInternalServerError)
		return
	}

	_, err = db.Exec(r.Context(), `
        INSERT INTO messages (from_id, to_id, content, created_at)
        VALUES ($1, $2, $3, NOW())
    `, fromID, toID, req.Content)
	if err != nil {
		http.Error(w, "メッセージ送信失敗", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("送信成功"))
}

func main() {
	ctx := context.Background()
	var err error
	db, err = pgx.Connect(ctx, "postgres://user:password@localhost:5432/chat_app_db")
	if err != nil {
		log.Fatal("DB接続失敗:", err)
	}
	defer db.Close(ctx)

	http.Handle("/signup", withCORS(http.HandlerFunc(signupHandler)))
	http.Handle("/login", withCORS(http.HandlerFunc(loginHandler)))
	http.Handle("/users", withCORS(http.HandlerFunc(usersHandler)))
	http.Handle("/messages", withCORS(http.HandlerFunc(messagesHandler))) // GET
	http.Handle("/send", withCORS(http.HandlerFunc(sendMessageHandler)))  // POST

	log.Println("サーバー起動: http://localhost:8081")
	log.Fatal(http.ListenAndServe(":8081", nil))
}
