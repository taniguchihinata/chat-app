package utils

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret = []byte("your_secret_key") // 本番環境では環境変数で管理！

// JWTを生成
func GenerateJWT(username string) (string, error) {
	claims := jwt.MapClaims{
		"username": username,
		"exp":      time.Now().Add(time.Hour * 24).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func ParseJWTFromRequest(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", errors.New("Authorizationヘッダーがありません")
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", errors.New("Authorizationヘッダーの形式が不正です")
	}

	tokenStr := parts[1]

	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if err != nil || !token.Valid {
		return "", errors.New("トークンが無効です")
	}

	username, ok := claims["username"].(string)
	if !ok {
		return "", errors.New("トークンにusernameが含まれていません")
	}

	return username, nil
}

// JWTを検証してユーザー名を取り出す
func ParseJWT(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if err != nil || !token.Valid {
		return "", err
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		username := claims["username"].(string)
		return username, nil
	}
	return "", err
}
