// internal/hub/hub.go
package hub

import (
	"chat-service/internal/model"
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

// 나중에 ingest-service gRPC 클라이언트가 이 인터페이스 구현
type IngestNotifier interface {
	NotifyChannelActive(ctx context.Context, channelID string, viewerCount int) error
}

// 채널별 클라이언트 관리
type Channel struct {
	clients map[*model.Client]bool
	mu      sync.RWMutex
}

type Hub struct {
	channels map[string]*Channel
	mu       sync.RWMutex

	register   chan *model.Client
	unregister chan *model.Client
	broadcast  chan *model.Message

	rdb *redis.Client
	// ingest-service 연동 포인트 (nil이면 스킵)
	ingestNotifier IngestNotifier
}

func NewHub(rdb *redis.Client, notifier IngestNotifier) *Hub {
	return &Hub{
		channels:       make(map[string]*Channel),
		register:       make(chan *model.Client, 256),
		unregister:     make(chan *model.Client, 256),
		broadcast:      make(chan *model.Message, 512),
		rdb:            rdb,
		ingestNotifier: notifier,
	}
}

func (h *Hub) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("hub stopped")
			return

		case client := <-h.register:
			h.addClient(ctx, client)

		case client := <-h.unregister:
			h.removeClient(ctx, client)

		case msg := <-h.broadcast:
			h.broadcastToChannel(msg)
		}
	}
}

func (h *Hub) Register(client *model.Client) {
	h.register <- client
}

func (h *Hub) Unregister(client *model.Client) {
	h.unregister <- client
}

func (h *Hub) Broadcast(msg *model.Message) {
	h.broadcast <- msg
}

func (h *Hub) addClient(ctx context.Context, client *model.Client) {
	h.mu.Lock()
	if _, ok := h.channels[client.ChannelID]; !ok {
		h.channels[client.ChannelID] = &Channel{
			clients: make(map[*model.Client]bool),
		}
	}
	ch := h.channels[client.ChannelID]
	h.mu.Unlock()

	ch.mu.Lock()
	ch.clients[client] = true
	viewerCount := len(ch.clients)
	ch.mu.Unlock()

	log.Info().
		Str("user_id", client.UserID).
		Str("channel_id", client.ChannelID).
		Int("viewer_count", viewerCount).
		Msg("client joined")

	// ingest-service에 채널 활성 알림 (gRPC)
	if h.ingestNotifier != nil {
		go func() {
			if err := h.ingestNotifier.NotifyChannelActive(ctx, client.ChannelID, viewerCount); err != nil {
				log.Warn().Err(err).Str("channel_id", client.ChannelID).Msg("ingest notify failed")
			}
		}()
	}

	h.rdb.ZAdd(ctx, "channel:"+client.ChannelID+":users", redis.Z{
		Score:  float64(time.Now().Unix()),
		Member: client.UserID,
	})

	h.rdb.ZAdd(ctx, "channels:ranking:viewers", redis.Z{
		Score:  float64(viewerCount),
		Member: client.ChannelID,
	})

}

func (h *Hub) removeClient(ctx context.Context, client *model.Client) {
	h.mu.RLock()
	ch, ok := h.channels[client.ChannelID]
	h.mu.RUnlock()
	if !ok {
		return
	}

	ch.mu.Lock()
	delete(ch.clients, client)
	viewerCount := len(ch.clients)
	ch.mu.Unlock()

	close(client.Send)

	log.Info().
		Str("user_id", client.UserID).
		Str("channel_id", client.ChannelID).
		Int("viewer_count", viewerCount).
		Msg("client left")

	// 유저 제거
	h.rdb.ZRem(ctx, "channel:"+client.ChannelID+":users", client.UserID)

	// 채널에 아무도 없으면 정리
	if viewerCount == 0 {
		h.mu.Lock()
		delete(h.channels, client.ChannelID)
		h.mu.Unlock()
		log.Info().Str("channel_id", client.ChannelID).Msg("channel removed (empty)")
		h.rdb.ZRem(ctx, "channels:ranking:viewers", client.ChannelID)
	} else {
		h.rdb.ZAdd(ctx, "channels:ranking:viewers", redis.Z{
			Score:  float64(viewerCount),
			Member: client.ChannelID,
		})
	}

	// ingest-service에 시청자 수 변경 알림 (gRPC)
	if h.ingestNotifier != nil {
		go func() {
			if err := h.ingestNotifier.NotifyChannelActive(ctx, client.ChannelID, viewerCount); err != nil {
				log.Warn().Err(err).Str("channel_id", client.ChannelID).Msg("ingest notify failed")
			}
		}()
	}
}

// 채널 활성 유저 목록 (입장 순)
func (h *Hub) ChannelUsers(ctx context.Context, channelID string) []string {
	result, _ := h.rdb.ZRange(ctx, "channel:"+channelID+":users", 0, -1).Result()
	return result
}

// 인기 채널 TOP N
func (h *Hub) TopChannels(ctx context.Context, n int) []redis.Z {
	result, _ := h.rdb.ZRevRangeWithScores(ctx, "channels:ranking:viewers", 0, int64(n-1)).Result()
	return result
}

func (h *Hub) UserRank(ctx context.Context, channelID, userID string) int64 {
	rank, err := h.rdb.ZRank(ctx, "channel:"+channelID+":users", userID).Result()
	if err != nil {
		return -1
	}
	return rank + 1 // 0-based → 1-based
}

func (h *Hub) broadcastToChannel(msg *model.Message) {
	h.mu.RLock()
	ch, ok := h.channels[msg.ChannelID]
	h.mu.RUnlock()
	if !ok {
		return
	}

	data, err := json.Marshal(model.ServerMessage{
		Type:    "chat",
		Payload: msg,
	})
	if err != nil {
		log.Error().Err(err).Msg("message marshal failed")
		return
	}

	ch.mu.RLock()
	defer ch.mu.RUnlock()

	for client := range ch.clients {
		select {
		case client.Send <- data:
		default:
			// 버퍼 꽉 찬 클라이언트 → 끊기
			log.Warn().Str("user_id", client.UserID).Msg("client send buffer full, dropping")
			close(client.Send)
			delete(ch.clients, client)
		}
	}
}

// ViewerCount - Sorted Set 활용 포인트
func (h *Hub) ViewerCount(channelID string) int {
	h.mu.RLock()
	ch, ok := h.channels[channelID]
	h.mu.RUnlock()
	if !ok {
		return 0
	}
	ch.mu.RLock()
	defer ch.mu.RUnlock()
	return len(ch.clients)
}
