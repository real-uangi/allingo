package log

import (
	"encoding/json"
	"fmt"
	"sync/atomic"
)

type Format uint8

const (
	TextFormat Format = iota
	JSONFormat
)

var outputFormat atomic.Uint32

// SetFormat changes the process-wide output format. Text is the default.
// Select the format before starting the application so queued entries agree.
func SetFormat(format Format) error {
	if format != TextFormat && format != JSONFormat {
		return fmt.Errorf("unknown log format %d", format)
	}
	outputFormat.Store(uint32(format))
	return nil
}

func formatJSONEntry(entry Entry) ([]byte, error) {
	fields := make(map[string]any, len(entry.Fields)+6)
	for key, value := range entry.Fields {
		if err, ok := value.(error); ok {
			value = err.Error()
		}
		fields[key] = value
	}
	// These metadata names are reserved; ordinary structured fields stay at the
	// top level so existing stage/order_id/action log queries keep working.
	fields["time"] = entry.Time
	fields["level"] = [...]string{"PANIC", "FATAL", "ERROR", "WARN", "INFO", "DEBUG", "TRACE"}[entry.Level]
	fields["msg"] = entry.Message
	fields[FieldLoggerName] = entry.LoggerName
	fields[FieldGoId] = entry.GoID
	if entry.Err != nil {
		fields["error"] = entry.Err.Error()
	}
	encoded, err := json.Marshal(fields)
	if err != nil {
		return nil, err
	}
	return append(encoded, '\n'), nil
}
