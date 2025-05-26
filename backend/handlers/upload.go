package handlers

import (
	"backend/utils"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

func UploadImageHandler(w http.ResponseWriter, r *http.Request) {
	// JWT チェック
	_, err := utils.ParseJWTFromRequest(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// multipart/form-data 形式で画像を受け取る
	file, header, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "画像ファイルが必要です", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// 保存ディレクトリを用意
	uploadDir := "./uploads"
	if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		http.Error(w, "保存先ディレクトリの作成に失敗", http.StatusInternalServerError)
		return
	}

	// ファイルパスを作成して保存
	filename := filepath.Base(header.Filename)
	dstPath := filepath.Join(uploadDir, filename)
	dst, err := os.Create(dstPath)
	if err != nil {
		http.Error(w, "ファイル保存に失敗しました", http.StatusInternalServerError)
		return
	}
	defer dst.Close()
	if _, err := io.Copy(dst, file); err != nil {
		http.Error(w, "保存中にエラーが発生しました", http.StatusInternalServerError)
		return
	}

	// クライアントにファイルURLを返す
	fileURL := fmt.Sprintf("/uploads/%s", filename)
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(fmt.Sprintf(`{"file_url":"%s"}`, fileURL)))
}
