package ws

import (
	"crypto/sha256"
	"encoding/base64"
	"time"
)

// Function returns the current time as a string in the format hh:mm:ss
func CurrentTime() string {
	return time.Now().Format("15:04:05")
}

// Hashes a string and returns its string
func HashPassword(pass string) string {
	h := sha256.New()
	h.Write([]byte(pass))
	return base64Hash(h.Sum(nil))
}

// Encodes a hash as a Base 64 string
func base64Hash(h []byte) string {
	return base64.URLEncoding.EncodeToString(h)
}

// Absolute function for ints
func Abs(x int) int {
	if x < 0 {
		return -1 * x
	}
	return x
}