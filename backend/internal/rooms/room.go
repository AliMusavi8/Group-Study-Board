package rooms

import (
	"context"
	"encoding/json"
	"sync"
	"sync/atomic"
	"time"

	"groupstudyboard/internal/config"
	"groupstudyboard/internal/db"
	"groupstudyboard/internal/models"
)

type Room struct {
	ID string

	store *db.Store
	cfg   config.Config

	register   chan *Client
	unregister chan *Client
	broadcast  chan models.Event

	clients map[*Client]bool

	mu           sync.RWMutex
	lastActivity time.Time
	seq          int64
	eventCount   int64
}

func NewRoom(id string, store *db.Store, cfg config.Config) *Room {
	return &Room{
		ID:         id,
		store:      store,
		cfg:        cfg,
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan models.Event, 256),
		clients:    make(map[*Client]bool),
		lastActivity: time.Now(),
	}
}

func (r *Room) Run(ctx context.Context) {
	_ = r.store.EnsureRoom(ctx, r.ID)
	for {
		select {
		case client := <-r.register:
			r.mu.Lock()
			r.clients[client] = true
			r.mu.Unlock()
			r.touch()
		case client := <-r.unregister:
			r.mu.Lock()
			if _, ok := r.clients[client]; ok {
				delete(r.clients, client)
				close(client.send)
			}
			r.mu.Unlock()
			r.touch()
		case event := <-r.broadcast:
			r.touch()
			payload, _ := json.Marshal(models.EventMessage{Type: "event", Event: event})
			r.mu.Lock()
			for client := range r.clients {
				select {
				case client.send <- payload:
				default:
					delete(r.clients, client)
					close(client.send)
				}
			}
			r.mu.Unlock()
		}
	}
}

func (r *Room) NextSeq() int64 {
	return atomic.AddInt64(&r.seq, 1)
}

func (r *Room) MaybeSnapshot() {
	if r.cfg.SnapshotEvery <= 0 {
		return
	}
	count := atomic.AddInt64(&r.eventCount, 1)
	if count%int64(r.cfg.SnapshotEvery) != 0 {
		return
	}
	go func(roomID string) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = r.store.CreateSnapshot(ctx, roomID)
	}(r.ID)
}

func (r *Room) ClientCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.clients)
}

func (r *Room) touch() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.lastActivity = time.Now()
}

func (r *Room) IsStale() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return time.Since(r.lastActivity) > r.cfg.RoomTTL
}

func (r *Room) Close() {
	r.mu.Lock()
	defer r.mu.Unlock()
	for client := range r.clients {
		close(client.send)
		delete(r.clients, client)
	}
}

func (r *Room) Register(client *Client) {
	r.register <- client
}

func (r *Room) Unregister(client *Client) {
	r.unregister <- client
}
