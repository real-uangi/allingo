package log

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func TestJSONOutputPreservesStructuredFieldsAndFiresHookOnce(t *testing.T) {
	resetHooksForTest()
	if err := SetFormat(JSONFormat); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { ExitTimeout(1); _ = SetFormat(TextFormat); resetHooksForTest() })
	var output bytes.Buffer
	logger := NewStdLogger("license.pay17")
	logger.SetOutput(&output)
	entries := make(chan Entry, 2)
	AddHook([]Level{ErrorLevel}, func(entry Entry) error { entries <- entry; return nil })
	logger.WithFields(Fields{"stage": "payment", "order_id": "order-123", "action": "notify", "count": 2}).Error(errors.New("transaction failed"), "payment failed")
	ExitTimeout(1)
	lines := strings.Split(strings.TrimSpace(output.String()), "\n")
	if len(lines) != 1 {
		t.Fatalf("expected one JSON record, got %d", len(lines))
	}
	var record map[string]any
	if err := json.Unmarshal([]byte(lines[0]), &record); err != nil {
		t.Fatal(err)
	}
	if record["stage"] != "payment" || record["order_id"] != "order-123" || record["action"] != "notify" || record["count"] != float64(2) || record["error"] != "transaction failed" || record["msg"] != "payment failed" || record["level"] != "ERROR" || record[FieldLoggerName] != "license.pay17" || record["time"] == nil || record[FieldGoId] == nil {
		t.Fatalf("JSON output lost log context: %#v", record)
	}
	entry := receiveEntry(t, entries)
	if entry.Fields["order_id"] != "order-123" {
		t.Fatal("structured hook lost its fields")
	}
	select {
	case <-entries:
		t.Fatal("hook executed twice")
	default:
	}
}

func TestSetFormatRejectsUnknownValue(t *testing.T) {
	if err := SetFormat(Format(255)); err == nil {
		t.Fatal("expected invalid format error")
	}
}
