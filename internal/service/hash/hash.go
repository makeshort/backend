package hash

import (
	"crypto/sha256"
	"fmt"
)

type Hasher struct {
	salt string
}

// New returns a new Hasher instance with given salt.
func New(salt string) *Hasher {
	return &Hasher{salt: salt}
}

// Create creates a hashed string from given string.
func (h *Hasher) Create(s string) string {
	hash := sha256.New()
	hash.Write([]byte(s + h.salt))
	sum := hash.Sum(nil)

	return fmt.Sprintf("%x", sum)
}
