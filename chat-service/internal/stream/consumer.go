// internal/stream/consumer.go
package stream

import (
	"chat-service/internal/config"
	"chat-service/internal/model"
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

type MessageBroadcaster interface {
	Broadcast(msg *model.Message)
}

type Consumer struct {
	rdb         *redis.Client
	cfg         *config.Config
	broadcaster MessageBroadcaster
}

func NewConsumer(rdb *redis.Client, cfg *config.Config, broadcaster MessageBroadcaster) *Consumer {
	return &Consumer{
		rdb:         rdb,
		cfg:         cfg,
		broadcaster: broadcaster,
	}
}

func (c *Consumer) Start(ctx context.Context) {
	c.ensureGroup(ctx)
	go c.consume(ctx)
	go c.reclaimLoop(ctx)
	log.Info().
		Str("stream", c.cfg.StreamChatMsg).
		Str("consumer", c.cfg.ConsumerName).
		Msg("chat consumer started")
}

// ── 소비 ─────────────────────────────────────────────────────
// 메시지 꺼내자마자 goroutine으로 던져서 블로킹 없음
// 채팅은 순서 보장보다 지연 최소화가 우선

func (c *Consumer) consume(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		streams, err := c.rdb.XReadGroup(ctx, &redis.XReadGroupArgs{
			Group:    c.cfg.GroupChat,
			Consumer: c.cfg.ConsumerName,
			Streams:  []string{c.cfg.StreamChatMsg, ">"},
			Count:    50, // 한 번에 많이 꺼내서 Redis 왕복 줄이기
			Block:    5 * time.Second,
		}).Result()

		if err != nil {
			if ctx.Err() != nil {
				return
			}
			if err == redis.Nil {
				continue
			}
			if isNoGroupError(err) {
				c.ensureGroup(ctx)
			} else {
				log.Error().Err(err).Msg("xreadgroup error")
				time.Sleep(time.Second)
			}
			continue
		}

		for _, stream := range streams {
			for _, msg := range stream.Messages {
				// goroutine으로 던지고 바로 다음 메시지
				go c.handle(ctx, msg)
			}
		}
	}
}

func (c *Consumer) handle(ctx context.Context, msg redis.XMessage) {
	payload, ok := msg.Values["payload"].(string)
	if !ok {
		log.Error().Str("msg_id", msg.ID).Msg("payload missing")
		c.ack(ctx, msg.ID)
		return
	}

	var chatMsg model.Message
	if err := json.Unmarshal([]byte(payload), &chatMsg); err != nil {
		log.Error().Err(err).Msg("unmarshal failed")
		c.ack(ctx, msg.ID)
		return
	}

	chatMsg.MessageID = msg.ID
	c.broadcaster.Broadcast(&chatMsg)
	c.ack(ctx, msg.ID)
}

// ── PEL 재처리 루프 ───────────────────────────────────────────
// 채팅은 PEL 중요도 낮음 → 1분 주기로 가볍게

func (c *Consumer) reclaimLoop(ctx context.Context) {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.reclaim(ctx)
		}
	}
}

func (c *Consumer) reclaim(ctx context.Context) {
	cursor := "0-0"
	for {
		msgs, next, err := c.rdb.XAutoClaim(ctx, &redis.XAutoClaimArgs{
			Stream:   c.cfg.StreamChatMsg,
			Group:    c.cfg.GroupChat,
			Consumer: c.cfg.ConsumerName,
			MinIdle:  2 * time.Minute,
			Start:    cursor,
			Count:    100,
		}).Result()

		if err != nil {
			log.Error().Err(err).Msg("reclaim failed")
			return
		}

		for _, msg := range msgs {
			go c.handle(ctx, msg)
		}

		if next == "0-0" || len(msgs) == 0 {
			break
		}
		cursor = next
	}
}

func (c *Consumer) ensureGroup(ctx context.Context) {
	err := c.rdb.XGroupCreateMkStream(ctx, c.cfg.StreamChatMsg, c.cfg.GroupChat, "$").Err()
	if err != nil && err.Error() != "BUSYGROUP Consumer Group name already exists" {
		log.Error().Err(err).Msg("xgroup create failed")
	}
}

func (c *Consumer) ack(ctx context.Context, msgID string) {
	if err := c.rdb.XAck(ctx, c.cfg.StreamChatMsg, c.cfg.GroupChat, msgID).Err(); err != nil {
		log.Error().Err(err).Str("msg_id", msgID).Msg("xack failed")
	}
}

func isNoGroupError(err error) bool {
	return strings.Contains(err.Error(), "NOGROUP")
}
