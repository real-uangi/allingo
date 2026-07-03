package random

import (
	"fmt"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestString(t *testing.T) {
	for i := 0; i < 10; i++ {
		s := String(32)
		if len(s) != 32 {
			t.Fatalf("unexpected length: got %d, want 32", len(s))
		}
		if !isValidSecureString(s) {
			t.Fatalf("invalid characters in %q", s)
		}
		t.Log(s)
	}
}

func TestFastString(t *testing.T) {
	for i := 0; i < 10; i++ {
		s := FastString(32)
		if len(s) != 32 {
			t.Fatalf("unexpected length: got %d, want 32", len(s))
		}
		if !isValidSecureString(s) {
			t.Fatalf("invalid characters in %q", s)
		}
		t.Log(s)
	}
}

func TestSecureString(t *testing.T) {
	for i := 0; i < 10; i++ {
		s := SecureString(32)
		if len(s) != 32 {
			t.Fatalf("unexpected length: got %d, want 32", len(s))
		}
		if !isValidSecureString(s) {
			t.Fatalf("invalid characters in %q", s)
		}
		t.Log(s)
	}
}

func BenchmarkFastString(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = FastString(32)
	}
}

func BenchmarkSecureString(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = SecureString(32)
	}
}

func BenchmarkUUID(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = uuid.NewString()
	}
}

func TestStringLengths(t *testing.T) {
	cases := []int{0, 1, 16, 32}
	for _, n := range cases {
		t.Run(fmt.Sprintf("FastString/%d", n), func(t *testing.T) {
			s := FastString(n)
			if len(s) != n {
				t.Fatalf("unexpected length: got %d, want %d", len(s), n)
			}
			if !isValidSecureString(s) {
				t.Fatalf("invalid characters in %q", s)
			}
		})
		t.Run(fmt.Sprintf("SecureString/%d", n), func(t *testing.T) {
			s := SecureString(n)
			if len(s) != n {
				t.Fatalf("unexpected length: got %d, want %d", len(s), n)
			}
			if !isValidSecureString(s) {
				t.Fatalf("invalid characters in %q", s)
			}
		})
	}
}

func TestSecureLetters(t *testing.T) {
	if len(secureLetters) != 64 {
		t.Fatalf("secureLetters length = %d, want 64", len(secureLetters))
	}
	for _, c := range secureLetters {
		if !strings.ContainsRune(baseLetters, c) {
			t.Fatalf("secureLetters contains invalid character %q", c)
		}
	}
	for _, c := range baseLetters {
		if !strings.ContainsRune(secureLetters, c) {
			t.Fatalf("secureLetters missing base character %q", c)
		}
	}
	t.Logf("secureLetters = %q", secureLetters)
}

func TestBuildSecureLetters(t *testing.T) {
	cases := []struct {
		name          string
		char1, char2  byte
		pos1, pos2    int
		wantPositions []int // expected positions of char1 and char2 in result
	}{
		{"ordered", 'X', 'Y', 5, 10, []int{5, 10}},
		{"reversed", 'X', 'Y', 10, 5, []int{5, 10}},
		{"same position", 'X', 'Y', 7, 7, []int{7, 8}},
		{"same position at end", 'X', 'Y', 62, 62, []int{62, 63}},
		{"at start", 'X', 'Y', 0, 1, []int{0, 1}},
		{"at end", 'X', 'Y', 62, 63, []int{62, 63}},
		{"one at end", 'X', 'Y', 30, 63, []int{30, 63}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := buildSecureLettersWith(tc.char1, tc.char2, tc.pos1, tc.pos2)
			if len(s) != 64 {
				t.Fatalf("length = %d, want 64", len(s))
			}
			if s[tc.wantPositions[0]] != tc.char1 {
				t.Fatalf("char1 at position %d = %q, want %q", tc.wantPositions[0], s[tc.wantPositions[0]], tc.char1)
			}
			if s[tc.wantPositions[1]] != tc.char2 {
				t.Fatalf("char2 at position %d = %q, want %q", tc.wantPositions[1], s[tc.wantPositions[1]], tc.char2)
			}
		})
	}
}

func TestFastStringNegativeLengthPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for negative length")
		}
	}()
	_ = FastString(-1)
}

func TestSecureStringNegativeLengthPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for negative length")
		}
	}()
	_ = SecureString(-1)
}

func isValidSecureString(s string) bool {
	for _, c := range s {
		if !strings.ContainsRune(secureLetters, c) {
			return false
		}
	}
	return true
}
