// internal/stream/producer.go
package stream

import (
	"chat-service/internal/config"
	"chat-service/internal/model"
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

type Producer struct {
	rdb *redis.Client
	cfg *config.Config
}

func NewProducer(rdb *redis.Client, cfg *config.Config) *Producer {
	return &Producer{rdb: rdb, cfg: cfg}
}

// 채팅 메시지 → Redis Stream publish
func (p *Producer) PublishChatMessage(ctx context.Context, msg *model.Message) error {
	msg.CreatedAt = time.Now()

	payload, err := json.Marshal(msg)
	if err != nil {
		log.Error().Err(err).Str("channel_id", msg.ChannelID).Msg("message marshal failed")
		return err
	}

	id, err := p.rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: p.cfg.StreamChatMsg,
		MaxLen: 10000, // 최대 10000개 유지 (오래된 메시지 자동 삭제)
		Approx: true,  // ~ 붙여서 성능 최적화
		Values: map[string]interface{}{
			"channel_id": msg.ChannelID,
			"payload":    string(payload),
		},
	}).Result()

	if err != nil {
		log.Error().Err(err).Str("channel_id", msg.ChannelID).Msg("xadd failed")
		return err
	}

	log.Debug().
		Str("stream_id", id).
		Str("channel_id", msg.ChannelID).
		Str("user_id", msg.UserID).
		Msg("chat message published")

	return nil
}
