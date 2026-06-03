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
	"io"
	"strings"
	"sync/atomic"
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

	ExitTimeout(1)
}

func TestExitTimeoutFlushesQueuedLogs(t *testing.T) {
	var output bytes.Buffer
	logger := NewStdLogger("allingo.common.log.application.TestLogger")
	logger.SetOutput(&output)

	logger.Info("flush before exit")
	ExitTimeout(1)

	if !strings.Contains(output.String(), "flush before exit") {
		t.Fatalf("expected queued log to be flushed, got %q", output.String())
	}
}

func TestWithFieldSendsFieldsToHooksButNotConsole(t *testing.T) {
	resetHooksForTest()
	t.Cleanup(resetHooksForTest)

	var output bytes.Buffer
	logger := NewStdLogger("allingo.common.log.application.TestLogger")
	logger.SetOutput(&output)

	entries := make(chan Entry, 1)
	AddHook([]Level{InfoLevel}, func(entry Entry) error {
		entries <- entry
		return nil
	})

	logger.WithField("secret", "field-value").Info("hello")
	ExitTimeout(1)

	entry := receiveEntry(t, entries)
	if entry.Fields["secret"] != "field-value" {
		t.Fatalf("expected hook field, got %#v", entry.Fields)
	}
	if strings.Contains(output.String(), "field-value") {
		t.Fatalf("expected structured field to stay out of console output, got %q", output.String())
	}
}

func TestChainedWithFieldAccumulatesFieldsOnWrapper(t *testing.T) {
	resetHooksForTest()
	t.Cleanup(resetHooksForTest)

	logger := NewStdLogger("allingo.common.log.application.TestLogger")
	logger.SetOutput(io.Discard)

	entries := make(chan Entry, 1)
	AddHook([]Level{InfoLevel}, func(entry Entry) error {
		entries <- entry
		return nil
	})

	logger.WithField("task", "sync").WithField("attempt", 3).Info("start")
	ExitTimeout(1)

	entry := receiveEntry(t, entries)
	if entry.Fields["task"] != "sync" || entry.Fields["attempt"] != 3 {
		t.Fatalf("expected chained fields, got %#v", entry.Fields)
	}
}

func TestPlainLoggerDoesNotCarryFieldsFromPreviousWrapper(t *testing.T) {
	resetHooksForTest()
	t.Cleanup(resetHooksForTest)

	logger := NewStdLogger("allingo.common.log.application.TestLogger")
	logger.SetOutput(io.Discard)

	entries := make(chan Entry, 2)
	AddHook([]Level{InfoLevel}, func(entry Entry) error {
		entries <- entry
		return nil
	})

	logger.WithField("request_id", "abc").Info("with field")
	logger.Info("without field")
	ExitTimeout(1)

	first := receiveEntry(t, entries)
	second := receiveEntry(t, entries)
	if first.Fields["request_id"] != "abc" {
		t.Fatalf("expected first entry field, got %#v", first.Fields)
	}
	if len(second.Fields) != 0 {
		t.Fatalf("expected no leaked fields, got %#v", second.Fields)
	}
}

func TestWithFieldsSnapshotsCallerMap(t *testing.T) {
	resetHooksForTest()
	t.Cleanup(resetHooksForTest)

	logger := NewStdLogger("allingo.common.log.application.TestLogger")
	logger.SetOutput(io.Discard)

	entries := make(chan Entry, 1)
	AddHook([]Level{InfoLevel}, func(entry Entry) error {
		entries <- entry
		return nil
	})

	fields := Fields{"status": "queued"}
	logger.WithFields(fields).Info("job")
	fields["status"] = "mutated"
	ExitTimeout(1)

	entry := receiveEntry(t, entries)
	if entry.Fields["status"] != "queued" {
		t.Fatalf("expected snapshot field, got %#v", entry.Fields)
	}
}

func TestInterfaceWorksWithStdLoggerAndLogWrapper(t *testing.T) {
	resetHooksForTest()
	t.Cleanup(resetHooksForTest)

	logger := NewStdLogger("allingo.common.log.application.TestLogger")
	logger.SetOutput(io.Discard)

	entries := make(chan Entry, 2)
	AddHook([]Level{InfoLevel}, func(entry Entry) error {
		entries <- entry
		return nil
	})

	var base Interface = logger
	base.WithField("source", "logger-interface").Info("from logger interface")

	var wrapper Interface = logger.WithField("source", "wrapper-interface")
	wrapper.WithField("extra", true).Infof("from %s", "wrapper interface")
	ExitTimeout(1)

	first := receiveEntry(t, entries)
	if first.Message != "from logger interface" {
		t.Fatalf("expected logger interface message, got %q", first.Message)
	}
	if first.Fields["source"] != "logger-interface" {
		t.Fatalf("expected logger interface field, got %#v", first.Fields)
	}

	second := receiveEntry(t, entries)
	if second.Message != "from wrapper interface" {
		t.Fatalf("expected wrapper interface message, got %q", second.Message)
	}
	if second.Fields["source"] != "wrapper-interface" || second.Fields["extra"] != true {
		t.Fatalf("expected wrapper interface fields, got %#v", second.Fields)
	}
}

func TestHookLevelFiltering(t *testing.T) {
	resetHooksForTest()
	t.Cleanup(resetHooksForTest)

	logger := NewStdLogger("allingo.common.log.application.TestLogger")
	logger.SetOutput(io.Discard)

	entries := make(chan Entry, 2)
	AddHook([]Level{InfoLevel}, func(entry Entry) error {
		entries <- entry
		return nil
	})

	logger.Warn("ignored")
	logger.Info("accepted")
	ExitTimeout(1)

	entry := receiveEntry(t, entries)
	if entry.Level != InfoLevel || entry.Message != "accepted" {
		t.Fatalf("expected only info hook entry, got level=%d message=%q", entry.Level, entry.Message)
	}
	assertNoEntry(t, entries)
}

func TestErrorfEntryIncludesErrorAndFieldsForHooks(t *testing.T) {
	resetHooksForTest()
	t.Cleanup(resetHooksForTest)

	logger := NewStdLogger("allingo.common.log.application.TestLogger")
	logger.SetOutput(io.Discard)

	entries := make(chan Entry, 1)
	AddHook([]Level{ErrorLevel}, func(entry Entry) error {
		entries <- entry
		return nil
	})

	err := errors.New("boom")
	logger.WithField("op", "save").Errorf(err, "failed %s", "save")
	ExitTimeout(1)

	entry := receiveEntry(t, entries)
	if entry.Level != ErrorLevel {
		t.Fatalf("expected error level, got %d", entry.Level)
	}
	if entry.Message != "failed save" {
		t.Fatalf("expected formatted message, got %q", entry.Message)
	}
	if entry.Err != err {
		t.Fatalf("expected original error, got %v", entry.Err)
	}
	if entry.Fields["op"] != "save" {
		t.Fatalf("expected error field, got %#v", entry.Fields)
	}
}

func TestHookErrorsAndPanicsDoNotStopLaterHooks(t *testing.T) {
	resetHooksForTest()
	t.Cleanup(resetHooksForTest)

	logger := NewStdLogger("allingo.common.log.application.TestLogger")
	logger.SetOutput(io.Discard)

	ran := make(chan struct{}, 1)
	AddHook([]Level{InfoLevel}, func(entry Entry) error {
		return errors.New("hook failed")
	})
	AddHook([]Level{InfoLevel}, func(entry Entry) error {
		panic("hook panicked")
	})
	AddHook([]Level{InfoLevel}, func(entry Entry) error {
		ran <- struct{}{}
		return nil
	})

	logger.Info("hook test")
	ExitTimeout(1)

	select {
	case <-ran:
	default:
		t.Fatal("expected hook after error and panic to run")
	}
	if HookErrorCount() != 2 {
		t.Fatalf("expected two hook errors, got %d", HookErrorCount())
	}
}

func TestExitTimeoutFlushesHooks(t *testing.T) {
	resetHooksForTest()
	t.Cleanup(resetHooksForTest)

	logger := NewStdLogger("allingo.common.log.application.TestLogger")
	logger.SetOutput(io.Discard)

	var called atomic.Int64
	AddHook([]Level{InfoLevel}, func(entry Entry) error {
		called.Add(1)
		return nil
	})

	logger.Info("flush hook")
	ExitTimeout(1)

	if called.Load() != 1 {
		t.Fatalf("expected hook to run before ExitTimeout returns, got %d", called.Load())
	}
}

func TestLowLevelLogsDropWhenQueueIsFull(t *testing.T) {
	resetHooksForTest()
	t.Cleanup(resetHooksForTest)

	logger := NewStdLogger("allingo.common.log.application.TestLogger")
	logger.SetOutput(io.Discard)

	started := make(chan struct{})
	release := make(chan struct{})
	var blocked atomic.Bool
	AddHook([]Level{InfoLevel}, func(entry Entry) error {
		if blocked.CompareAndSwap(false, true) {
			close(started)
		}
		<-release
		return nil
	})

	logger.Info("block worker")
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("expected hook to block worker")
	}

	before := DroppedCount()
	for i := 0; i < cap(logQueue); i++ {
		logger.Info("fill")
	}
	logger.Info("drop")
	after := DroppedCount()

	close(release)
	ExitTimeout(1)

	if after <= before {
		t.Fatalf("expected dropped count to increase, before=%d after=%d", before, after)
	}
}

func receiveEntry(t *testing.T, entries <-chan Entry) Entry {
	t.Helper()
	select {
	case entry := <-entries:
		return entry
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for hook entry")
		return Entry{}
	}
}

func assertNoEntry(t *testing.T, entries <-chan Entry) {
	t.Helper()
	select {
	case entry := <-entries:
		t.Fatalf("expected no extra hook entry, got %#v", entry)
	default:
	}
}
