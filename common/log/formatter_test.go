package log

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func TestFormatterAppendsErrorSuffix(t *testing.T) {
	output := renderEntry(ErrorLevel, "request failed", errors.New("boom"))
	if !strings.Contains(output, "request failed | err=boom\n") {
		t.Fatalf("expected error suffix in output, got %q", output)
	}
}

func TestFormatterAvoidsDuplicateErrorSuffixWhenMessageAlreadyContainsError(t *testing.T) {
	output := renderEntry(ErrorLevel, "request failed: boom", errors.New("boom"))
	if strings.Count(output, "boom") != 1 {
		t.Fatalf("expected error text once, got %q", output)
	}
	if strings.Contains(output, "| err=boom") {
		t.Fatalf("expected no duplicated error suffix, got %q", output)
	}
}

func TestFormatterKeepsErrorOnlyMessageSingle(t *testing.T) {
	output := renderEntry(ErrorLevel, "boom", errors.New("boom"))
	if strings.Count(output, "boom") != 1 {
		t.Fatalf("expected error text once, got %q", output)
	}
}

func TestFormatterLeavesNonErrorMessageUnchanged(t *testing.T) {
	output := renderEntry(InfoLevel, "hello", nil)
	if !strings.Contains(output, "hello\n") {
		t.Fatalf("expected info message in output, got %q", output)
	}
	if strings.Contains(output, "err=") {
		t.Fatalf("expected non-error output to stay unchanged, got %q", output)
	}
}

func renderEntry(level Level, message string, err error) string {
	entry := Entry{
		Time:       time.Date(2026, 4, 9, 12, 0, 0, 0, time.UTC),
		Level:      level,
		LoggerName: "allingo.common.log.test",
		GoID:       7,
		Message:    message,
		Err:        err,
	}
	return string(formatEntry(entry, newMiddleInfos(entry.LoggerName)))
}
