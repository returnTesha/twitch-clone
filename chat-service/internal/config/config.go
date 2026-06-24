// internal/config/config.go
package config

import (
	"os"
	"strconv"
)

type Config struct {
	// 서버
	Port string

	// Redis
	RedisAddr     string
	RedisPassword string
	RedisDB       int

	// Redis Stream
	StreamChatMsg string // 채팅 메시지 스트림
	GroupChat     string // consumer group
	ConsumerName  string // pod hostname 기반

	// JWT
	JWTSecret string
}

func Load() *Config {
	return &Config{
		Port:          getEnv("PORT", "8080"),
		RedisAddr:     getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		RedisDB:       getEnvInt("REDIS_DB", 0),

		StreamChatMsg: getEnv("STREAM_CHAT_MSG", "chat:msg"),
		GroupChat:     getEnv("GROUP_CHAT", "chat-group"),
		ConsumerName:  getEnv("CONSUMER_NAME", mustHostname()),

		JWTSecret: getEnv("JWT_SECRET", "local-secret"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}

func mustHostname() string {
	h, err := os.Hostname()
	if err != nil {
		return "consumer-1"
	}
	return "consumer-" + h
}
