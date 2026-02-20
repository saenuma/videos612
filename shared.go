package l8f

import (
	"math/rand"
	"os"
)

func doesPathExists(p string) bool {
	if _, err := os.Stat(p); os.IsNotExist(err) {
		return false
	}
	return true
}

func untestedRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz1234567890"

	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}
