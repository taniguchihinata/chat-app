package handlers

import (
	"backend/utils"
	"context"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/jackc/pgx/v5/pgxpool"
)

// WebSocketで使用する構造体
type WSMessage struct {
	Type      string `json:"type"` //"message" or "read"
	RoomID    int    `json:"room_id"`
	Text      string `json:"text"`
	Image     string `json:"image,omitempty"`
	Sender    int    `json:"sender_id"`
	Username  string `json:"username"`
	MessageID int    `json:"message_id,omitempty"` // 既読通知用
}

type Client struct {
	Conn   *websocket.Conn
	RoomID int
}

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}

	roomClients = make(map[int][]*Client) // ルームごとのクライアントリスト
	mu          sync.Mutex                // goroutine競合防止
)

// handlers/ws.go の下部
var RoomClients = roomClients
var RoomClientsMutex = &mu

func WebSocketHandler(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenStr := r.URL.Query().Get("token")
		if tokenStr == "" {
			http.Error(w, "トークンが指定されていません", http.StatusUnauthorized)
			return
		}

		username, err := utils.ParseJWT(tokenStr)
		if err != nil {
			http.Error(w, "トークンが無効です", http.StatusUnauthorized)
			return
		}
		log.Printf("WebSocket認証成功: %s", username)

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("WebSocketアップグレード失敗:", err)
			return
		}
		defer conn.Close()

		var initMsg WSMessage
		if err := conn.ReadJSON(&initMsg); err != nil {
			log.Println("初期メッセージ読み込み失敗:", err)
			return
		}
		client := &Client{Conn: conn, RoomID: initMsg.RoomID}

		mu.Lock()
		roomClients[client.RoomID] = append(roomClients[client.RoomID], client)
		mu.Unlock()

		log.Printf("ルーム %d にクライアント接続", client.RoomID)

		for {
			var msg WSMessage
			if err := conn.ReadJSON(&msg); err != nil {
				log.Println("メッセージ読み込みエラー:", err)
				break
			}

			switch msg.Type {
			case "leave":
				log.Printf("ユーザー %s がルーム %d を離れました", msg.Username, msg.RoomID)
				mu.Lock()
				for _, c := range roomClients[msg.RoomID] {
					_ = c.Conn.WriteJSON(msg)
				}
				mu.Unlock()
				continue

			case "read":
				log.Printf("既読通知: %s がメッセージ %d を読んだ（ルーム %d）", msg.Username, msg.MessageID, msg.RoomID)

				var readerID int
				err := db.QueryRow(context.Background(),
					"SELECT id FROM users WHERE username=$1", msg.Username,
				).Scan(&readerID)
				if err == nil {
					_, err = db.Exec(context.Background(),
						`INSERT INTO message_reads (message_id, user_id)
						 VALUES ($1, $2)
						 ON CONFLICT DO NOTHING`,
						msg.MessageID, readerID)
					if err != nil {
						log.Println("既読情報保存失敗:", err)
					}
				} else {
					log.Println("既読ユーザーID取得失敗:", err)
				}

				mu.Lock()
				for _, c := range roomClients[msg.RoomID] {
					_ = c.Conn.WriteJSON(msg)
				}
				mu.Unlock()
				continue

			case "message":
				log.Printf("ルーム%d: %s", msg.RoomID, msg.Text)

				var senderID int
				err := db.QueryRow(context.Background(),
					"SELECT id FROM users WHERE username = $1", username,
				).Scan(&senderID)
				if err != nil {
					log.Println("sender_id取得失敗:", err)
					continue
				}

				originalText := msg.Text

				err = db.QueryRow(context.Background(),
					`INSERT INTO messages (room_id, sender_id, text, created_at)
					 VALUES ($1, $2, $3, now()) RETURNING id`,
					msg.RoomID, senderID, msg.Text,
				).Scan(&msg.MessageID)
				if err != nil {
					log.Println("メッセージ保存失敗:", err)
					continue
				}

				msg.Username = username
				msg.Text = originalText

				mu.Lock()
				for _, c := range roomClients[msg.RoomID] {
					_ = c.Conn.WriteJSON(msg)
				}
				mu.Unlock()
			}
		}

		mu.Lock()
		clients := roomClients[client.RoomID]
		for i, c := range clients {
			if c == client {
				roomClients[client.RoomID] = append(clients[:i], clients[i+1:]...)
				break
			}
		}
		mu.Unlock()

		log.Println("WebSocket切断")
	}
}
