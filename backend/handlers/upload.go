package handlers

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func UploadAttachmentHandler(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// クエリからメッセージID取得
		messageIDStr := r.URL.Query().Get("message_id")
		messageID, err := strconv.Atoi(messageIDStr)
		if err != nil || messageID <= 0 {
			http.Error(w, "message_idが無効", http.StatusBadRequest)
			return
		}

		// ファイル取得
		file, header, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "ファイルの読み込みに失敗しました", http.StatusBadRequest)
			return
		}
		defer file.Close()

		// 保存先パス生成
		dir := "uploads"
		os.MkdirAll(dir, os.ModePerm)
		filename := fmt.Sprintf("%d_%s", time.Now().UnixNano(), header.Filename)
		path := filepath.Join(dir, filename)

		// 保存処理
		dst, err := os.Create(path)
		if err != nil {
			http.Error(w, "ファイル保存失敗", http.StatusInternalServerError)
			return
		}
		defer dst.Close()
		_, err = io.Copy(dst, file)
		if err != nil {
			http.Error(w, "ファイル書き込み失敗", http.StatusInternalServerError)
			return
		}

		// DBに登録
		_, err = db.Exec(context.Background(),
			`INSERT INTO message_attachments (message_id, file_url) VALUES ($1, $2)`,
			messageID, "/uploads/"+filename)
		if err != nil {
			http.Error(w, "DB登録失敗", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("アップロード成功"))
	}
}
