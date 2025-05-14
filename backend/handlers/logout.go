package handlers

import (
	"backend/utils"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
)

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Authorizationヘッダーがありません", http.StatusUnauthorized)
		return
	}
	tokenStr := strings.Split(authHeader, " ")[1]

	claims := jwt.MapClaims{}
	_, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return utils.GetJWTSecret(), nil
	})
	if err != nil {
		http.Error(w, "トークンが不正です", http.StatusUnauthorized)
		return
	}

	jti := claims["jti"].(string)
	expUnix := int64(claims["exp"].(float64))
	ttl := time.Until(time.Unix(expUnix, 0))

	err = utils.RedisClient.Set(utils.Ctx, "blacklist:"+jti, "1", ttl).Err()
	if err != nil {
		http.Error(w, "トークン失効に失敗しました", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ログアウトしました"))
}
