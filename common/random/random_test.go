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
		if !isValidString(s) {
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
		if !isValidString(s) {
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
		if !isValidString(s) {
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
			if !isValidString(s) {
				t.Fatalf("invalid characters in %q", s)
			}
		})
		t.Run(fmt.Sprintf("SecureString/%d", n), func(t *testing.T) {
			s := SecureString(n)
			if len(s) != n {
				t.Fatalf("unexpected length: got %d, want %d", len(s), n)
			}
			if !isValidString(s) {
				t.Fatalf("invalid characters in %q", s)
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

func isValidString(s string) bool {
	for _, c := range s {
		if !strings.ContainsRune(stringLetters, c) {
			return false
		}
	}
	return true
}
