package main

import (
    "encoding/json"
    "log"
    "net/http"
    "golang.org/x/crypto/bcrypt" //ハッシュ化
    "context"//データベース追加処理
    "github.com/jackc/pgx/v5"//データベース追加処理
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

    // ハッシュ化
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
    if err != nil {
        http.Error(w, "パスワードのハッシュ化に失敗しました", http.StatusInternalServerError)
        return
    }

    //INSERT処理
    _, err = db.Exec(r.Context(), `
        INSERT INTO users (username, password_hash)
        VALUES ($1, $2)
    `, req.Username, string(hashedPassword))

    if err != nil {
        log.Println("INSERT失敗:",err)
        http.Error(w, "ユーザー登録に失敗しました", http.StatusInternalServerError)
        return
    }


    log.Printf("登録ユーザー: %s, ハッシュ化パスワード: %s", req.Username, string(hashedPassword))

    w.WriteHeader(http.StatusOK)
    w.Write([]byte("ユーザー登録成功"))
}

//CORS対応ミドルウェア
func withCORS(handler http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Access-Control-Allow-Origin", "*")
        w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

        if r.Method == http.MethodOptions {
            w.WriteHeader(http.StatusOK)
            return
        }

        handler.ServeHTTP(w, r)
    })
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
    log.Println("サーバー起動: http://localhost:8081")
    log.Fatal(http.ListenAndServe(":8081", nil))
}
