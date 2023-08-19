package alias

import (
	"math/rand"
	"time"
)

const chars = "abcdefghijklmnopqrstuvwxuz0123456789"

// Generate generates an random chars string given length. Great compatible for alias.
func Generate(size int) string {
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	chars := []rune(chars)

	b := make([]rune, size)
	for i := range b {
		b[i] = chars[rnd.Intn(len(chars))]
	}

	return string(b)
}

// TODO: Change math/rand, mb hash ???