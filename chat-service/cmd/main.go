// cmd/main.go
package main

import (
	"chat-service/internal/auth"
	"chat-service/internal/config"
	"chat-service/internal/handler"
	"chat-service/internal/hub"
	"chat-service/internal/stream"
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	// 로거
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	// 설정
	cfg := config.Load()

	// Redis
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})
	defer rdb.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Redis 연결 확인
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatal().Err(err).Msg("redis connection failed")
	}
	log.Info().Str("addr", cfg.RedisAddr).Msg("redis connected")

	// Hub (ingest-service 붙을 때 nil → gRPC 클라이언트로 교체)
	h := hub.NewHub(rdb, nil)
	go h.Run(ctx)

	// Stream
	producer := stream.NewProducer(rdb, cfg)
	consumer := stream.NewConsumer(rdb, cfg, h)
	consumer.Start(ctx)

	// Auth
	jwtManager := auth.NewJWTManager(cfg)

	// Handler
	wsHandler := handler.NewWSHandler(h, producer)

	// Router
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(gin.Logger())

	// 헬스체크
	r.GET("/health", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// 토큰 발급 (개발/테스트용)
	r.POST("/token", func(ctx *gin.Context) {
		var req struct {
			UserID   string `json:"user_id"`
			Username string `json:"username"`
		}
		if err := ctx.ShouldBindJSON(&req); err != nil || req.UserID == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "user_id required"})
			return
		}
		token, err := jwtManager.Generate(req.UserID, req.Username)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "token generation failed"})
			return
		}
		ctx.JSON(http.StatusOK, gin.H{"token": token})
	})

	// 채널 활성 유저 목록
	r.GET("/channels/:channelId/users", func(ctx *gin.Context) {
		channelID := ctx.Param("channelId")
		users := h.ChannelUsers(ctx.Request.Context(), channelID)
		ctx.JSON(http.StatusOK, gin.H{
			"channel_id": channelID,
			"users":      users,
			"count":      len(users),
		})
	})

	// 인기 채널 TOP 10
	r.GET("/channels/ranking", func(ctx *gin.Context) {
		channels := h.TopChannels(ctx.Request.Context(), 10)
		ctx.JSON(http.StatusOK, gin.H{
			"ranking": channels,
		})
	})

	// 내 입장 순번
	r.GET("/channels/:channelId/rank/:userId", func(ctx *gin.Context) {
		channelID := ctx.Param("channelId")
		userID := ctx.Param("userId")
		rank := h.UserRank(ctx.Request.Context(), channelID, userID)
		if rank == -1 {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "user not in channel"})
			return
		}
		ctx.JSON(http.StatusOK, gin.H{
			"channel_id": channelID,
			"user_id":    userID,
			"rank":       rank,
		})
	})
	// WebSocket
	ws := r.Group("/ws")
	ws.Use(jwtManager.Middleware())
	ws.GET("/:channelId", wsHandler.Handle)

	// 서버
	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
	}

	// graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Info().Str("port", cfg.Port).Msg("chat service started")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("server error")
		}
	}()

	<-quit
	log.Info().Msg("shutting down...")

	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error().Err(err).Msg("shutdown error")
	}

	log.Info().Msg("chat service stopped")
}
