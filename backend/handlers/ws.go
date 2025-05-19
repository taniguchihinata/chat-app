package handlers

import (
	"context"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/jackc/pgx/v5/pgxpool"
)

// WebSocketで使用する構造体
type WSMessage struct {
	RoomID int    `json:"room_id"`
	Text   string `json:"text"`
	Sender int    `json:"sender_id"`
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

func WebSocketHandler(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("WebSocketアップグレード失敗:", err)
			return
		}
		defer conn.Close()

		// 初期メッセージを読み込み、クライアントのRoomIDを確定
		var initMsg WSMessage
		if err := conn.ReadJSON(&initMsg); err != nil {
			log.Println("初期メッセージ読み込み失敗:", err)
			return
		}
		client := &Client{Conn: conn, RoomID: initMsg.RoomID}

		// クライアントをルームに追加
		mu.Lock()
		roomClients[client.RoomID] = append(roomClients[client.RoomID], client)
		mu.Unlock()

		log.Printf("ルーム %d にクライアント接続", client.RoomID)

		// メッセージ受信ループ
		for {
			var msg WSMessage
			if err := conn.ReadJSON(&msg); err != nil {
				log.Println("メッセージ読み込みエラー:", err)
				break
			}

			log.Printf("ルーム%d: %s", msg.RoomID, msg.Text)

			// メッセージをデータベースに保存
			_, err := db.Exec(context.Background(),
				`INSERT INTO messages (room_id, sender_id, text, created_at)
				 VALUES ($1, $2, $3, now())`,
				msg.RoomID, msg.Sender, msg.Text,
			)
			if err != nil {
				log.Println("メッセージ保存失敗:", err)
			}

			// ブロードキャスト（同じルームの全クライアントに送信）
			mu.Lock()
			for _, c := range roomClients[msg.RoomID] {
				if err := c.Conn.WriteJSON(msg); err != nil {
					log.Println("送信エラー:", err)
				}
			}
			mu.Unlock()
		}

		// 切断されたクライアントをルームから削除
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
