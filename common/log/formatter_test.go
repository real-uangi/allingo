package log

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
)

func TestFormatterAppendsErrorSuffix(t *testing.T) {
	output := formatEntry(t, logrus.ErrorLevel, "request failed", errors.New("boom"))
	if !strings.Contains(output, "request failed | err=boom\n") {
		t.Fatalf("expected error suffix in output, got %q", output)
	}
}

func TestFormatterAvoidsDuplicateErrorSuffixWhenMessageAlreadyContainsError(t *testing.T) {
	output := formatEntry(t, logrus.ErrorLevel, "request failed: boom", errors.New("boom"))
	if strings.Count(output, "boom") != 1 {
		t.Fatalf("expected error text once, got %q", output)
	}
	if strings.Contains(output, "| err=boom") {
		t.Fatalf("expected no duplicated error suffix, got %q", output)
	}
}

func TestFormatterKeepsErrorOnlyMessageSingle(t *testing.T) {
	output := formatEntry(t, logrus.ErrorLevel, "boom", errors.New("boom"))
	if strings.Count(output, "boom") != 1 {
		t.Fatalf("expected error text once, got %q", output)
	}
}

func TestFormatterLeavesNonErrorMessageUnchanged(t *testing.T) {
	output := formatEntry(t, logrus.InfoLevel, "hello", nil)
	if !strings.Contains(output, "hello\n") {
		t.Fatalf("expected info message in output, got %q", output)
	}
	if strings.Contains(output, "err=") {
		t.Fatalf("expected non-error output to stay unchanged, got %q", output)
	}
}

func formatEntry(t *testing.T, level logrus.Level, message string, err error) string {
	t.Helper()

	entry := &logrus.Entry{
		Logger:  logrus.New(),
		Data:    logrus.Fields{FieldGoId: int64(7)},
		Time:    time.Date(2026, 4, 9, 12, 0, 0, 0, time.UTC),
		Level:   level,
		Message: message,
	}
	if err != nil {
		entry.Data[logrus.ErrorKey] = err
	}

	output, formatErr := newFormatter("allingo.common.log.test").Format(entry)
	if formatErr != nil {
		t.Fatalf("format entry: %v", formatErr)
	}
	return string(output)
}
