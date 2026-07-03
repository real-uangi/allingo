package random

import (
	crand "crypto/rand"
	"math/rand"
	"sync"
	"time"
	"unsafe"
)

const stringLetters = "abcdefghijklmnopqrstuvwxyz1234567890ABCDEFGHIJKLMNOPQRSTUVWXYZ"

const (
	// 6 bits to represent a letter index
	letterIdBits = 6
	// All 1-bits as many as letterIdBits
	letterIdMask = 1<<letterIdBits - 1
	letterIdMax  = 63 / letterIdBits
)

var rngPool = sync.Pool{
	New: func() any {
		return rand.New(rand.NewSource(time.Now().UnixNano()))
	},
}

// FastString generates a random string of length n using math/rand.
// It is fast but NOT cryptographically secure. Suitable for things like trace IDs.
func FastString(n int) string {
	src := rngPool.Get().(*rand.Rand)
	defer rngPool.Put(src)
	b := make([]byte, n)
	// A rand.Int63() generates 63 random bits, enough for letterIdMax letters!
	for i, cache, remain := n-1, src.Int63(), letterIdMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdMax
		}
		if idx := int(cache & letterIdMask); idx < len(stringLetters) {
			b[i] = stringLetters[idx]
			i--
		}
		cache >>= letterIdBits
		remain--
	}
	return *(*string)(unsafe.Pointer(&b))
}

// String is kept for backward compatibility.
//
// Deprecated: Use FastString for non-cryptographic purposes.
// For security-sensitive use cases, use SecureString instead.
func String(n int) string {
	return FastString(n)
}

// SecureString generates a cryptographically secure random string of length n.
// It uses crypto/rand and rejection sampling to ensure a uniform distribution.
// Use this for passwords, tokens, verification codes, and other security-sensitive values.
func SecureString(n int) string {
	b := make([]byte, n)
	buf := make([]byte, n)
	var pos int
	for i := 0; i < n; i++ {
		for {
			if pos == 0 {
				if _, err := crand.Read(buf); err != nil {
					panic(err)
				}
				pos = len(buf)
			}
			pos--
			idx := int(buf[pos])
			if idx < len(stringLetters) {
				b[i] = stringLetters[idx]
				break
			}
		}
	}
	return *(*string)(unsafe.Pointer(&b))
}
