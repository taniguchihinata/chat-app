// サーバーシステム
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
	utils.InitRedis()
	ctx := context.Background()
	var err error
	db, err := pgxpool.New(ctx, "postgres://user:password@localhost:5432/chat_app_db")
	if err != nil {
		log.Fatal("DB接続失敗:", err)
	}
	defer db.Close()

	//withCORSの中に書いてある関数が動いている感じ
	http.Handle("/signup", utils.WithCORS(handlers.SignupHandler(db)))
	http.Handle("/login", utils.WithCORS(handlers.LoginHandler(db)))
	http.Handle("/users", utils.WithCORS(handlers.UsersHandler(db)))
	http.Handle("/messages", utils.WithCORS(handlers.GetMessagesHandler(db))) // GET
	http.Handle("/send", utils.WithCORS(handlers.SendMessageHandler(db)))     // POST
	http.Handle("/rooms", utils.WithCORS(handlers.RoomsHandler(db)))
	http.HandleFunc("/me", utils.WithCORS(handlers.MeHandler))
	http.HandleFunc("/logout", utils.WithCORS(handlers.LogoutHandler))
	http.HandleFunc("/ws", handlers.WebSocketHandler(db))

	log.Println("サーバー起動: http://localhost:8081")
	log.Fatal(http.ListenAndServe(":8081", nil))
}
