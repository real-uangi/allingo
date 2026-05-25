/*
 * Copyright © 2024 Uangi. All rights reserved.
 * @author uangi
 * @date 2024/11/7 16:57
 */

// Package log
package log

import (
	"bytes"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	logger := NewStdLogger("allingo.common.log.application.TestLogger")
	logger.Error(errors.New("error"), "error")
	logger.Warn("warn")
	logger.Info("info")
	logger.Debug("debug")
	logger.Trace("trace")

	logger.Print("print ggg\n\n")
	logger.Print("print ggg\n")

	time.Sleep(2 * time.Second)

}

func TestExitTimeoutFlushesQueuedLogs(t *testing.T) {
	var output bytes.Buffer
	logger := NewStdLogger("allingo.common.log.application.TestLogger")
	logger.logger.SetOutput(&output)

	logger.Info("flush before exit")
	ExitTimeout(1)

	if !strings.Contains(output.String(), "flush before exit") {
		t.Fatalf("expected queued log to be flushed, got %q", output.String())
	}
}
