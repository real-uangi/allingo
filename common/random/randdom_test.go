package random

import (
	"github.com/google/uuid"
	"testing"
)

func TestString(t *testing.T) {
	for i := 0; i < 10; i++ {
		t.Log(String(32))
	}
}

func BenchmarkGenerate(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = String(32)
	}
}

func BenchmarkUUID(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = uuid.NewString()
	}
}
