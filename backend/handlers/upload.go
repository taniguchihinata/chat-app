package handlers

import (
	"backend/utils"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func randomString(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "rand"
	}
	return hex.EncodeToString(b)
}

func UploadHandler(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 認証（トークンが有効かチェック）
		username, err := utils.ParseJWTFromRequest(r)
		if err != nil {
			log.Println("JWT認証エラー:", err)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		log.Println("アップロードユーザー:", username)

		// multipart/form-data をパース
		err = r.ParseMultipartForm(10 << 20) // 最大10MB
		if err != nil {
			http.Error(w, "ファイルサイズが大きすぎます", http.StatusBadRequest)
			return
		}

		file, handler, err := r.FormFile("image")
		if err != nil {
			http.Error(w, "ファイル読み込み失敗", http.StatusBadRequest)
			return
		}
		defer file.Close()

		// 保存先ディレクトリ作成（なければ）
		saveDir := "./uploads"
		if err := os.MkdirAll(saveDir, os.ModePerm); err != nil {
			log.Println("フォルダ作成失敗:", err)
			http.Error(w, "サーバー内部エラー", http.StatusInternalServerError)
			return
		}

		// 拡張子・ファイル名の安全処理
		ext := filepath.Ext(handler.Filename)
		base := strings.TrimSuffix(handler.Filename, ext)
		safeName := strings.ReplaceAll(base, " ", "_")
		uniqueName := fmt.Sprintf("%s_%d_%s%s", safeName, time.Now().UnixNano(), randomString(4), ext)
		savePath := filepath.Join(saveDir, uniqueName)

		// ファイル保存
		dst, err := os.Create(savePath)
		if err != nil {
			log.Println("ファイル作成失敗:", err)
			http.Error(w, "保存に失敗しました", http.StatusInternalServerError)
			return
		}
		defer dst.Close()

		if _, err := io.Copy(dst, file); err != nil {
			log.Println("ファイル保存失敗:", err)
			http.Error(w, "コピーに失敗しました", http.StatusInternalServerError)
			return
		}

		// 成功レスポンス
		fileURL := fmt.Sprintf("/uploads/%s", uniqueName)
		log.Printf("アップロード完了: %s", fileURL)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"file_url": fileURL,
		})
	}
}
