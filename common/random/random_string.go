package random

import (
	crand "crypto/rand"
	"math/rand"
	"sync"
	"time"
	"unsafe"
)

// baseLetters is the 62-character URL-safe alphabet used as the foundation for
// the runtime-generated 64-character alphabet.
const baseLetters = "abcdefghijklmnopqrstuvwxyz1234567890ABCDEFGHIJKLMNOPQRSTUVWXYZ"

// secureLetters is the runtime-generated 64-character alphabet.
//
// Why a 64-character alphabet?
//
// baseLetters has 62 characters, which is not a power of two. Mapping random
// bytes to it requires rejection sampling, wasting ~24% of entropy reads and
// adding branches. By filling two extra slots with randomly chosen characters
// at runtime, the table becomes 64 entries long and each random byte can be
// mapped with a single "& 0x3F" operation.
//
// Performance trade-off (Apple M4, n=32):
//   - 62-char + per-byte rejection sampling: ~6 026 ns/op for SecureString
//   - 62-char + batched rejection sampling: ~1 114 ns/op for SecureString
//   - 64-char runtime-filled table: ~206 ns/op for SecureString
//   - FastString stays roughly the same at ~50 ns/op
//
// Uniformity trade-off:
//   - The two filler characters occur with probability 2/64 instead of 1/62.
//   - Entropy per character drops from log2(62) ≈ 5.954 bits to ~5.938 bits,
//     a loss of ~0.3%. The exact filler characters and positions vary per
//     process instance.
var secureLetters string

func init() {
	secureLetters = buildSecureLetters()
}

// buildSecureLetters constructs a 64-character alphabet from baseLetters by
// inserting two distinct filler characters at random positions. The fillers are
// chosen from baseLetters without replacement so exactly two slots are doubled.
func buildSecureLetters() string {
	// Choose two distinct characters from baseLetters.
	idx1 := randByteMax(len(baseLetters))
	idx2 := randByteMax(len(baseLetters) - 1)
	if idx2 >= idx1 {
		idx2++
	}
	char1 := baseLetters[idx1]
	char2 := baseLetters[idx2]

	// Choose two insertion positions in the final 64-character alphabet.
	// If they collide, pick a single position that leaves room for both chars.
	pos1 := randByteMax(64)
	pos2 := randByteMax(64)
	if pos1 == pos2 {
		pos1 = randByteMax(63)
		pos2 = pos1
	}

	return buildSecureLettersWith(char1, char2, pos1, pos2)
}

// buildSecureLettersWith constructs a 64-character alphabet from baseLetters by
// inserting char1 and char2 at the given positions within the final 64-character
// result. If pos1 > pos2 they are swapped before insertion.
func buildSecureLettersWith(char1, char2 byte, pos1, pos2 int) string {
	if pos1 > pos2 {
		pos1, pos2 = pos2, pos1
	}

	b := make([]byte, 64)
	if pos1 == pos2 {
		copy(b[:pos1], baseLetters[:pos1])
		b[pos1] = char1
		b[pos1+1] = char2
		copy(b[pos1+2:], baseLetters[pos1:])
		return string(b)
	}

	copy(b[:pos1], baseLetters[:pos1])
	b[pos1] = char1
	copy(b[pos1+1:pos2], baseLetters[pos1:pos2-1])
	b[pos2] = char2
	copy(b[pos2+1:], baseLetters[pos2-1:])
	return string(b)
}

// randByteMax returns a uniformly distributed random integer in [0, max).
// It uses crypto/rand and rejection sampling.
func randByteMax(max int) int {
	if max <= 0 || max > 256 {
		panic("randByteMax: max must be in (0, 256]")
	}
	limit := 256 - (256 % max)
	var buf [1]byte
	for {
		if _, err := crand.Read(buf[:]); err != nil {
			panic(err)
		}
		v := int(buf[0])
		if v < limit {
			return v % max
		}
	}
}

var rngPool = sync.Pool{
	New: func() any {
		return rand.New(rand.NewSource(time.Now().UnixNano()))
	},
}

// FastString generates a random string of length n using math/rand.
// It is fast but NOT cryptographically secure. Suitable for things like trace IDs.
// It uses the runtime-generated 64-character URL-safe alphabet for fast bit mapping.
func FastString(n int) string {
	src := rngPool.Get().(*rand.Rand)
	defer rngPool.Put(src)
	b := make([]byte, n)
	//nolint:errcheck // math/rand.Read never returns an error for a non-nil Rand.
	src.Read(b)
	for i := range b {
		b[i] = secureLetters[b[i]&0x3F]
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
// It uses crypto/rand and the runtime-generated 64-character URL-safe alphabet.
//
// The 64-character table trades a small amount of uniformity for a large speed
// gain: compared to the original 62-character alphabet with per-byte rejection
// sampling, this implementation is roughly 29x faster on the benchmark hardware.
// The final switch from batched rejection sampling to the 64-character table
// accounts for about a 5.4x speedup. The two runtime-selected filler characters
// appear slightly more often than the rest (~0.3% entropy loss per character);
// the randomness source itself remains cryptographic.
// Use this for passwords, tokens, verification codes, and other security-sensitive values.
func SecureString(n int) string {
	b := make([]byte, n)
	if _, err := crand.Read(b); err != nil {
		panic(err)
	}
	for i := range b {
		b[i] = secureLetters[b[i]&0x3F]
	}
	return *(*string)(unsafe.Pointer(&b))
}
