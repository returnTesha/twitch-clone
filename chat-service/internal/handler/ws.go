// internal/handler/ws.go
package handler

import (
	"chat-service/internal/hub"
	"chat-service/internal/model"
	"chat-service/internal/stream"
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512 // bytes
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // 개발 환경 - 프로덕션에서는 origin 체크
	},
}

type WSHandler struct {
	hub      *hub.Hub
	producer *stream.Producer
}

func NewWSHandler(h *hub.Hub, p *stream.Producer) *WSHandler {
	return &WSHandler{hub: h, producer: p}
}

// GET /ws/:channelId
func (wh *WSHandler) Handle(ctx *gin.Context) {
	channelID := ctx.Param("channelId")
	if channelID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "channelId required"})
		return
	}

	// JWT 미들웨어에서 set한 값
	userID := ctx.GetString("user_id")
	username := ctx.GetString("username")

	conn, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		log.Error().Err(err).Msg("websocket upgrade failed")
		return
	}

	client := &model.Client{
		ID:        uuid.New().String(),
		UserID:    userID,
		Username:  username,
		ChannelID: channelID,
		Send:      make(chan []byte, 256),
	}

	wh.hub.Register(client)

	// 입장 알림
	wh.notifyJoin(ctx.Request.Context(), client)

	// read / write 분리
	go wh.writePump(conn, client)
	wh.readPump(ctx.Request.Context(), conn, client)
}

// ── Read Pump ─────────────────────────────────────────────────
// 클라이언트 → 서버 메시지 수신

func (wh *WSHandler) readPump(ctx context.Context, conn *websocket.Conn, client *model.Client) {
	defer func() {
		wh.hub.Unregister(client)
		wh.notifyLeave(ctx, client)
		conn.Close()
	}()

	conn.SetReadLimit(maxMessageSize)
	conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, raw, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err,
				websocket.CloseGoingAway,
				websocket.CloseAbnormalClosure,
			) {
				log.Warn().Err(err).Str("user_id", client.UserID).Msg("unexpected close")
			}
			break
		}

		var clientMsg model.ClientMessage
		if err := json.Unmarshal(raw, &clientMsg); err != nil {
			log.Warn().Err(err).Str("user_id", client.UserID).Msg("invalid message format")
			continue
		}

		// 다른 채널에 메시지 보내는 거 방지
		if clientMsg.ChannelID != client.ChannelID {
			log.Warn().
				Str("user_id", client.UserID).
				Str("client_channel", client.ChannelID).
				Str("msg_channel", clientMsg.ChannelID).
				Msg("channel mismatch, ignored")
			continue
		}

		msg := &model.Message{
			ChannelID: client.ChannelID,
			UserID:    client.UserID,
			Username:  client.Username,
			Content:   clientMsg.Content,
		}

		// Redis Stream으로 publish → consumer가 받아서 hub broadcast
		if err := wh.producer.PublishChatMessage(ctx, msg); err != nil {
			log.Error().Err(err).Str("user_id", client.UserID).Msg("publish failed")
		}
	}
}

// ── Write Pump ────────────────────────────────────────────────
// 서버 → 클라이언트 메시지 전송

func (wh *WSHandler) writePump(conn *websocket.Conn, client *model.Client) {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		conn.Close()
	}()

	for {
		select {
		case msg, ok := <-client.Send:
			conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				log.Warn().Err(err).Str("user_id", client.UserID).Msg("write failed")
				return
			}

		case <-ticker.C:
			conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// ── 입장/퇴장 알림 ────────────────────────────────────────────

func (wh *WSHandler) notifyJoin(ctx context.Context, client *model.Client) {
	msg := &model.Message{
		ChannelID: client.ChannelID,
		UserID:    client.UserID,
		Username:  client.Username,
		Content:   client.Username + " 님이 입장했습니다.",
	}
	if err := wh.producer.PublishChatMessage(ctx, msg); err != nil {
		log.Warn().Err(err).Msg("join notify failed")
	}
}

func (wh *WSHandler) notifyLeave(ctx context.Context, client *model.Client) {
	msg := &model.Message{
		ChannelID: client.ChannelID,
		UserID:    client.UserID,
		Username:  client.Username,
		Content:   client.Username + " 님이 퇴장했습니다.",
	}
	if err := wh.producer.PublishChatMessage(ctx, msg); err != nil {
		log.Warn().Err(err).Msg("leave notify failed")
	}
}
