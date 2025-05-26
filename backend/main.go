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
	http.Handle("/rooms/", utils.WithCORS(handlers.GetRoomDetailHandler(db)))
	http.Handle("/read", utils.WithCORS(handlers.MarkAsReadHandler(db)))
	http.Handle("/read_status", utils.WithCORS(handlers.GetReadStatusHandler(db)))
	http.Handle("/read_status_full", utils.WithCORS(handlers.GetFullReadStatusHandler(db)))
	http.Handle("/unread_count", utils.WithCORS(handlers.GetUnreadCountHandler(db)))
	http.Handle("/mark_all_read", utils.WithCORS(handlers.MarkAllAsReadHandler(db)))
	http.Handle("/room_info", utils.WithCORS(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			roomID := r.URL.Query().Get("room")
			r.URL.Path = "/rooms/" + roomID
		}
		handlers.GetRoomDetailHandler(db).ServeHTTP(w, r)
	})))
	http.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir("./uploads"))))
	http.Handle("/upload", utils.WithCORS(handlers.UploadImageHandler))

	log.Println("サーバー起動: http://localhost:8081")
	log.Fatal(http.ListenAndServe(":8081", nil))
}
