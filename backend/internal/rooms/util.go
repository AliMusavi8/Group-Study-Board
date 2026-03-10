package rooms

import (
	"crypto/rand"
)

const roomAlphabet = "23456789abcdefghjkmnpqrstuvwxyz"

func GenerateRoomID(length int) (string, error) {
	buf := make([]byte, length)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	for i := range buf {
		buf[i] = roomAlphabet[int(buf[i])%len(roomAlphabet)]
	}
	return string(buf), nil
}
