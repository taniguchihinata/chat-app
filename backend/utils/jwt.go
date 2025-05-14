package utils

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var jwtSecret = []byte("your_secret_key") // 本番環境では環境変数で管理！

// JWTを生成
func GenerateJWT(username string) (string, error) {
	jti := uuid.New().String()
	claims := jwt.MapClaims{
		"username": username,
		"exp":      time.Now().Add(time.Hour * 24).Unix(),
		"jti":      jti,
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

	// jti の存在確認
	jti, ok := claims["jti"].(string)
	if !ok || jti == "" {
		return "", errors.New("トークンに jti が含まれていません")
	}

	// Redis でブラックリスト確認
	exists, err := RedisClient.Exists(Ctx, "blacklist:"+jti).Result()
	if err != nil {
		return "", errors.New("Redis 確認中にエラーが発生しました: " + err.Error())
	}
	if exists == 1 {
		return "", errors.New("このトークンは無効化されています")
	}

	// username 抽出
	username, ok := claims["username"].(string)
	if !ok {
		return "", errors.New("トークンに username が含まれていません")
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

// jwt.go に追加（末尾などに）
func GetJWTSecret() []byte {
	return jwtSecret
}
