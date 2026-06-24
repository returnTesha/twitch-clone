// internal/model/message.go
package model

import "time"

// Redis Stream → hub → 클라이언트로 전달되는 채팅 메시지
type Message struct {
	MessageID string    `json:"message_id"` // Redis Stream ID
	ChannelID string    `json:"channel_id"`
	UserID    string    `json:"user_id"`
	Username  string    `json:"username"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

// WebSocket으로 클라이언트가 보내는 메시지
type ClientMessage struct {
	ChannelID string `json:"channel_id"`
	Content   string `json:"content"`
}

// WebSocket으로 클라이언트에게 보내는 메시지
type ServerMessage struct {
	Type    string   `json:"type"` // "chat" | "join" | "leave" | "error"
	Payload *Message `json:"payload"`
}

// Redis Stream에 저장할 때 쓰는 map
func (m *Message) ToStreamPayload() map[string]interface{} {
	return map[string]interface{}{
		"message_id": m.MessageID,
		"channel_id": m.ChannelID,
		"user_id":    m.UserID,
		"username":   m.Username,
		"content":    m.Content,
		"created_at": m.CreatedAt.Format(time.RFC3339),
	}
}

// Redis Stream에서 꺼낼 때 파싱
func FromStreamPayload(values map[string]interface{}) *Message {
	createdAt, _ := time.Parse(time.RFC3339, str(values["created_at"]))
	return &Message{
		MessageID: str(values["message_id"]),
		ChannelID: str(values["channel_id"]),
		UserID:    str(values["user_id"]),
		Username:  str(values["username"]),
		Content:   str(values["content"]),
		CreatedAt: createdAt,
	}
}

func str(v interface{}) string {
	if v == nil {
		return ""
	}
	s, _ := v.(string)
	return s
}
