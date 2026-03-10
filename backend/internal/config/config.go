package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port            string
	MongoURI        string
	DatabaseName    string
	CorsOrigin      string
	RoomTTL         time.Duration
	MaxParticipants int
	RateLimitPerSec int
	RateLimitBurst  int
	SnapshotEvery   int
}

func Load() Config {
	return Config{
		Port:            getEnv("PORT", "8080"),
		MongoURI:        getEnv("MONGODB_URI", "mongodb://localhost:27017"),
		DatabaseName:    getEnv("DATABASE_NAME", "group_study_board"),
		CorsOrigin:      getEnv("CORS_ORIGIN", "http://localhost:4200"),
		RoomTTL:         time.Duration(getInt("ROOM_TTL_MINUTES", 60)) * time.Minute,
		MaxParticipants: getInt("MAX_PARTICIPANTS", 24),
		RateLimitPerSec: getInt("RATE_LIMIT_PER_SEC", 60),
		RateLimitBurst:  getInt("RATE_LIMIT_BURST", 120),
		SnapshotEvery:   getInt("SNAPSHOT_EVERY", 200),
	}
}

func getEnv(key, def string) string {
	val := os.Getenv(key)
	if val == "" {
		return def
	}
	return val
}

func getInt(key string, def int) int {
	val := os.Getenv(key)
	if val == "" {
		return def
	}
	parsed, err := strconv.Atoi(val)
	if err != nil {
		return def
	}
	return parsed
}
