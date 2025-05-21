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

		//WebSocket接続開始
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("WebSocketアップグレード失敗:", err)
			return
		}
		defer conn.Close()

		// 初期メッセージを読み込み、クライアントのRoomIDを確定
		var initMsg WSMessage
		if err := conn.ReadJSON(&initMsg); err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Println("致命的な初期メッセージ読み込み失敗:", err)
			} else {
				log.Printf("初期接続が完了前に終了（無視）: %v", err)
			}
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

			//既読通知
			if msg.Type == "read" {
				log.Printf("既読通知: %s がメッセージ %d を読んだ（ルーム %d）", msg.Username, msg.MessageID, msg.RoomID)

				var readerID int
				err := db.QueryRow(context.Background(),
					"SELECT id FROM users WHERE username=$1", msg.Username,
				).Scan(&readerID)
				if err != nil {
					log.Println("既読ユーザーID取得失敗:", err)
				} else {
					_, err := db.Exec(context.Background(),
						`INSERT INTO message_reads (message_id, user_id)
						 VALUES ($1, $2)
						 ON CONFLICT DO NOTHING`,
						msg.MessageID, readerID,
					)
					if err != nil {
						log.Println("既読情報保存失敗:", err)
					}
				}

				mu.Lock()
				for _, c := range roomClients[msg.RoomID] {
					if err := c.Conn.WriteJSON(msg); err != nil {
						log.Println("既読通知送信エラー:", err)
					}
				}
				mu.Unlock()

				continue // 下の保存処理には行かない
			}

			log.Printf("ルーム%d: %s", msg.RoomID, msg.Text)

			var senderID int
			err := db.QueryRow(context.Background(),
				"SELECT id FROM users WHERE username = $1", username,
			).Scan(&senderID)
			if err != nil {
				log.Println("sender_id取得失敗:", err)
				continue
			}

			// メッセージをデータベースに保存
			_, err = db.Exec(context.Background(),
				`INSERT INTO messages (room_id, sender_id, text, created_at)
				 VALUES ($1, $2, $3, now())`,
				msg.RoomID, senderID, msg.Text,
			)
			if err != nil {
				log.Println("メッセージ保存失敗:", err)
				continue
			}

			// ブロードキャスト（同じルームの全クライアントに送信）
			msg.Username = username

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
