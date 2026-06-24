// internal/model/client.go
package model

type Client struct {
	ID        string
	UserID    string
	Username  string
	ChannelID string
	Send      chan []byte
}
