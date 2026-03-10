package rooms

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"groupstudyboard/internal/config"
	"groupstudyboard/internal/models"

	"github.com/gorilla/websocket"
	"golang.org/x/time/rate"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 64 * 1024
)

type Client struct {
	id      string
	room    *Room
	conn    *websocket.Conn
	send    chan []byte
	limiter *rate.Limiter
	cfg     config.Config
}

func NewClient(id string, room *Room, conn *websocket.Conn, cfg config.Config) *Client {
	return &Client{
		id:      id,
		room:    room,
		conn:    conn,
		send:    make(chan []byte, 256),
		limiter: rate.NewLimiter(rate.Limit(cfg.RateLimitPerSec), cfg.RateLimitBurst),
		cfg:     cfg,
	}
}

func (c *Client) ReadPump() {
	defer func() {
		c.room.Unregister(c)
		_ = c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		var msg models.ClientEvent
		if err := c.conn.ReadJSON(&msg); err != nil {
			break
		}

		if !c.limiter.Allow() {
			_ = c.sendError("rate_limit")
			continue
		}

		if !isAllowedType(msg.Type) {
			_ = c.sendError("invalid_event_type")
			continue
		}

		event := models.Event{
			RoomID:    c.room.ID,
			Type:      msg.Type,
			ClientID:  c.id,
			Point:     msg.Point,
			Color:     msg.Color,
			Thickness: msg.Thickness,
			Tool:      msg.Tool,
			Seq:       c.room.NextSeq(),
			ServerTs:  time.Now().UnixMilli(),
		}

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		_ = c.room.store.SaveEvent(ctx, event)
		cancel()

		c.room.MaybeSnapshot()
		c.room.broadcast <- event
	}
}

func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		_ = c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}
		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) sendError(message string) error {
	payload, _ := json.Marshal(models.ErrorMessage{Type: "error", Message: message})
	select {
	case c.send <- payload:
		return nil
	default:
		return nil
	}
}

func isAllowedType(t string) bool {
	switch t {
	case "strokeStart", "strokeMove", "strokeEnd":
		return true
	default:
		return false
	}
}

func UpgradeWebsocket(w http.ResponseWriter, r *http.Request) (*websocket.Conn, error) {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(req *http.Request) bool {
			return true
		},
	}
	return upgrader.Upgrade(w, r, nil)
}
