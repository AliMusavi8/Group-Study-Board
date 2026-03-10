package rooms

import (
	"context"
	"sync"
	"time"

	"groupstudyboard/internal/config"
	"groupstudyboard/internal/db"
)

type Hub struct {
	store *db.Store
	cfg   config.Config

	mu    sync.RWMutex
	rooms map[string]*Room
}

func NewHub(store *db.Store, cfg config.Config) *Hub {
	return &Hub{
		store: store,
		cfg:   cfg,
		rooms: make(map[string]*Room),
	}
}

func (h *Hub) GetOrCreate(roomID string) *Room {
	h.mu.RLock()
	room := h.rooms[roomID]
	h.mu.RUnlock()
	if room != nil {
		return room
	}

	h.mu.Lock()
	defer h.mu.Unlock()
	if existing := h.rooms[roomID]; existing != nil {
		return existing
	}

	room = NewRoom(roomID, h.store, h.cfg)
	h.rooms[roomID] = room
	go room.Run(context.Background())
	return room
}

func (h *Hub) CleanupLoop(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			h.cleanup()
		}
	}
}

func (h *Hub) cleanup() {
	h.mu.Lock()
	defer h.mu.Unlock()
	for id, room := range h.rooms {
		if room.IsStale() && room.ClientCount() == 0 {
			room.Close()
			delete(h.rooms, id)
		}
	}
}
