package main

import (
	"context"
	"log"
	"net/http"

	"backend/handlers"
	"backend/utils"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	ctx := context.Background()
	var err error
	db, err := pgxpool.New(ctx, "postgres://user:password@localhost:5432/chat_app_db")
	if err != nil {
		log.Fatal("DB接続失敗:", err)
	}
	defer db.Close()

	// DBを各ハンドラーパッケージに注入
	handlers.DB = db

	http.Handle("/signup", utils.WithCORS(http.HandlerFunc(handlers.SignupHandler)))
	http.Handle("/login", utils.WithCORS(http.HandlerFunc(handlers.LoginHandler)))
	http.Handle("/users", utils.WithCORS(http.HandlerFunc(handlers.UsersHandler)))
	http.Handle("/messages", utils.WithCORS(http.HandlerFunc(handlers.MessagesHandler)))
	http.Handle("/send", utils.WithCORS(http.HandlerFunc(handlers.SendMessageHandler)))

	log.Println("サーバー起動: http://localhost:8081")
	log.Fatal(http.ListenAndServe(":8081", nil))
}
