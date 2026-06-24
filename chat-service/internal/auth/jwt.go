// internal/auth/jwt.go
package auth

import (
	"chat-service/internal/config"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

type JWTManager struct {
	secret []byte
}

func NewJWTManager(cfg *config.Config) *JWTManager {
	return &JWTManager{secret: []byte(cfg.JWTSecret)}
}

// 토큰 발급 (테스트/로그인용)
func (j *JWTManager) Generate(userID, username string) (string, error) {
	claims := &Claims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.secret)
}

// 토큰 검증
func (j *JWTManager) Verify(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return j.secret, nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}

// Gin 미들웨어
func (j *JWTManager) Middleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// WebSocket 업그레이드 시 쿼리파라미터로도 토큰 받을 수 있게
		// ws://host/ws?token=xxx
		tokenStr := ctx.Query("token")
		if tokenStr == "" {
			// Authorization: Bearer xxx
			auth := ctx.GetHeader("Authorization")
			if !strings.HasPrefix(auth, "Bearer ") {
				ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
				return
			}
			tokenStr = strings.TrimPrefix(auth, "Bearer ")
		}

		claims, err := j.Verify(tokenStr)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		ctx.Set("user_id", claims.UserID)
		ctx.Set("username", claims.Username)
		ctx.Next()
	}
}
