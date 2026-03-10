package main

import (
	"context"
	"net/http"
	"strings"
	"time"

	"groupstudyboard/internal/config"
	"groupstudyboard/internal/db"
	"groupstudyboard/internal/models"
	"groupstudyboard/internal/rooms"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

func main() {
	cfg := config.Load()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	store, err := db.NewStore(ctx, cfg)
	if err != nil {
		panic(err)
	}

	hub := rooms.NewHub(store, cfg)
	go hub.CleanupLoop(context.Background())

	router := gin.Default()
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{cfg.CorsOrigin},
		AllowMethods:     []string{"GET", "POST", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type"},
		AllowCredentials: true,
	}))

	router.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	router.POST("/rooms", func(c *gin.Context) {
		roomID, err := rooms.GenerateRoomID(10)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "room_id_generation_failed"})
			return
		}
		if err := store.EnsureRoom(c.Request.Context(), roomID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "room_create_failed"})
			return
		}
		_ = hub.GetOrCreate(roomID)
		c.JSON(http.StatusCreated, gin.H{"roomId": roomID})
	})

	router.GET("/rooms/:id/ws", func(c *gin.Context) {
		roomID := c.Param("id")
		clientID := c.Query("clientId")
		if clientID == "" {
			clientID = "guest-" + strings.ReplaceAll(time.Now().Format("150405.000"), ".", "")
		}

		room := hub.GetOrCreate(roomID)
		if room.ClientCount() >= cfg.MaxParticipants {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "room_full"})
			return
		}

		upgrader := websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(req *http.Request) bool {
				origin := req.Header.Get("Origin")
				if origin == "" {
					return true
				}
				return origin == cfg.CorsOrigin
			},
		}

		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			return
		}

		client := rooms.NewClient(clientID, room, conn, cfg)

		snapshot, events, err := store.LoadSnapshotAndEvents(c.Request.Context(), roomID)
		if err == nil {
			_ = conn.WriteJSON(models.SnapshotMessage{
				Type:     "snapshot",
				RoomID:   roomID,
				Snapshot: snapshot,
				Events:   events,
			})
		}

		room.Register(client)
		go client.WritePump()
		go client.ReadPump()
	})

	_ = router.Run("0.0.0.0:" + cfg.Port)
}
