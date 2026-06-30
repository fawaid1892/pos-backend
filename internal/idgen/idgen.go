package idgen

import (
	"crypto/rand"
	"math/big"
	"sync"
)

var mu sync.Mutex

// Generate generates a cryptographically secure 16-digit ID (1000000000000000 - 9999999999999999)
func Generate() int64 {
	mu.Lock()
	defer mu.Unlock()
	max := big.NewInt(9000000000000000) // 9e15
	for {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			panic(err)
		}
		id := n.Int64() + 1000000000000000 // ensure 16 digits
		return id
	}
}
