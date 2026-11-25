package task

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"go.uber.org/fx/fxtest"
)

func waitTaskSignal(t *testing.T, signal <-chan struct{}, message string) {
	t.Helper()
	select {
	case <-signal:
	case <-time.After(time.Second):
		t.Fatal(message)
	}
}

func TestManagedSchedulingStartsWithFxAndUsesIndependentContext(t *testing.T) {
	lifecycle := fxtest.NewLifecycle(t)
	manager := NewManagedManager(lifecycle, newTaskTestKV())
	t.Cleanup(func() { _ = manager.Stop(context.Background()) })
	seen := make(chan error, 1)
	task, err := manager.AddContext("managed-start", "@every 1h", func(ctx context.Context) error { seen <- ctx.Err(); return nil })
	if err != nil {
		t.Fatal(err)
	}
	if !manager.c.Entry(task.id).Next.IsZero() {
		t.Fatal("managed constructor started scheduling")
	}
	startContext, cancelStart := context.WithCancel(t.Context())
	if err := lifecycle.Start(startContext); err != nil {
		t.Fatal(err)
	}
	cancelStart()
	if manager.c.Entry(task.id).Next.IsZero() {
		t.Fatal("Fx startup did not start scheduling")
	}
	manager.c.Entry(task.id).Job.Run()
	select {
	case err := <-seen:
		if err != nil {
			t.Fatalf("task inherited canceled startup context: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("task did not run")
	}
	if err := lifecycle.Stop(t.Context()); err != nil {
		t.Fatal(err)
	}
}

func TestStopCancelsAndWaitsForActualAsyncWork(t *testing.T) {
	manager := NewManager(newTaskTestKV())
	entered, canceled, finished, release := make(chan struct{}), make(chan struct{}), make(chan struct{}), make(chan struct{})
	var releaseOnce sync.Once
	t.Cleanup(func() { releaseOnce.Do(func() { close(release) }); _ = manager.Stop(context.Background()) })
	task, err := manager.AddContext("drain-async", "@every 1h", func(ctx context.Context) error {
		close(entered)
		<-ctx.Done()
		close(canceled)
		<-release
		close(finished)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	// This cron callback returns after submission, while its async work is still
	// running. Waiting only for cron.Stop would incorrectly report completion.
	manager.c.Entry(task.id).Job.Run()
	waitTaskSignal(t, entered, "task did not enter async work")
	stopContext, cancelStop := context.WithTimeout(t.Context(), 20*time.Millisecond)
	defer cancelStop()
	if err := manager.Stop(stopContext); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Stop with unfinished async task = %v, want deadline", err)
	}
	waitTaskSignal(t, canceled, "Stop did not cancel running task")
	select {
	case <-finished:
		t.Fatal("blocked task unexpectedly finished")
	default:
	}
	releaseOnce.Do(func() { close(release) })
	if err := manager.Stop(t.Context()); err != nil {
		t.Fatal(err)
	}
	waitTaskSignal(t, finished, "Stop returned before async callback finished")
	if _, err := manager.Add("after-stop", "@every 1h", func() error { return nil }); !errors.Is(err, ErrStopped) {
		t.Fatalf("Add after Stop = %v", err)
	}
}

func TestTaskDoesNotOverlapOrAdmitWorkAfterStop(t *testing.T) {
	manager := NewManager(newTaskTestKV())
	entered, release := make(chan struct{}), make(chan struct{})
	var count atomic.Int32
	var releaseOnce sync.Once
	t.Cleanup(func() { releaseOnce.Do(func() { close(release) }); _ = manager.Stop(context.Background()) })
	task, err := manager.Add("no-overlap", "@every 1h", func() error { count.Add(1); close(entered); <-release; return nil })
	if err != nil {
		t.Fatal(err)
	}
	job := manager.c.Entry(task.id).Job
	job.Run()
	waitTaskSignal(t, entered, "first invocation did not start")
	var submitters sync.WaitGroup
	for range 16 {
		submitters.Go(job.Run)
	}
	submitters.Wait()
	if count.Load() != 1 {
		t.Fatalf("overlapping invocation count = %d", count.Load())
	}
	stopContext, cancelStop := context.WithTimeout(t.Context(), 20*time.Millisecond)
	defer cancelStop()
	if err := manager.Stop(stopContext); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("legacy callback Stop = %v, want honest timeout", err)
	}
	releaseOnce.Do(func() { close(release) })
	if err := manager.Stop(t.Context()); err != nil {
		t.Fatal(err)
	}
	job.Run()
	if count.Load() != 1 {
		t.Fatal("stopped manager admitted another invocation")
	}
}

func TestTriggerSharesManagedCancellationAndDrain(t *testing.T) {
	lifecycle := fxtest.NewLifecycle(t)
	manager := NewManagedManager(lifecycle, newTaskTestKV())
	entered, canceled, release := make(chan struct{}), make(chan struct{}), make(chan struct{})
	var releaseOnce sync.Once
	t.Cleanup(func() { releaseOnce.Do(func() { close(release) }); _ = manager.Stop(context.Background()) })
	manager.MustAddContext("initial-cleanup", "@every 1h", func(ctx context.Context) error {
		close(entered)
		<-ctx.Done()
		close(canceled)
		<-release
		return nil
	})
	if err := manager.Trigger("initial-cleanup"); !errors.Is(err, ErrNotStarted) {
		t.Fatalf("Trigger before startup = %v", err)
	}
	if err := lifecycle.Start(t.Context()); err != nil {
		t.Fatal(err)
	}
	if err := manager.Trigger("missing"); !errors.Is(err, ErrTaskNotFound) {
		t.Fatalf("unknown Trigger = %v", err)
	}
	if err := manager.Trigger("initial-cleanup"); err != nil {
		t.Fatal(err)
	}
	waitTaskSignal(t, entered, "Trigger did not start its registered task")
	if err := manager.Trigger("initial-cleanup"); err != nil {
		t.Fatal(err)
	}
	stopContext, cancelStop := context.WithTimeout(t.Context(), 20*time.Millisecond)
	defer cancelStop()
	if err := manager.Stop(stopContext); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Stop did not track triggered async work: %v", err)
	}
	waitTaskSignal(t, canceled, "triggered task did not receive cancellation")
	releaseOnce.Do(func() { close(release) })
	if err := manager.Stop(t.Context()); err != nil {
		t.Fatal(err)
	}
	if err := manager.Trigger("initial-cleanup"); !errors.Is(err, ErrStopped) {
		t.Fatalf("Trigger after Stop = %v", err)
	}
}

func TestParentCancellationStopsAdmissionBeforeLifecycleStop(t *testing.T) {
	parent, cancelParent := context.WithCancel(t.Context())
	defer cancelParent()
	lifecycle := fxtest.NewLifecycle(t)
	manager := NewManagedManagerWithContext(lifecycle, newTaskTestKV(), parent)
	t.Cleanup(func() { _ = manager.Stop(context.Background()) })
	entered, canceled, finished := make(chan struct{}), make(chan struct{}), make(chan struct{})
	var invocations atomic.Int32
	manager.MustAddContext("runtime-owned", "@every 1h", func(ctx context.Context) error {
		defer close(finished)
		invocations.Add(1)
		close(entered)
		<-ctx.Done()
		close(canceled)
		return nil
	})
	if err := lifecycle.Start(t.Context()); err != nil {
		t.Fatal(err)
	}
	if err := manager.Trigger("runtime-owned"); err != nil {
		t.Fatal(err)
	}
	waitTaskSignal(t, entered, "runtime-owned task did not start")
	cancelParent()
	waitTaskSignal(t, canceled, "parent cancellation did not reach the running task")
	waitTaskSignal(t, finished, "canceled callback did not finish")
	if err := manager.Trigger("runtime-owned"); !errors.Is(err, ErrStopped) {
		t.Fatalf("canceled parent still admitted a trigger: %v", err)
	}
	if _, err := manager.AddContext("after-cancel", "@every 1h", func(context.Context) error { invocations.Add(1); return nil }); !errors.Is(err, ErrStopped) {
		t.Fatalf("canceled parent still admitted registration: %v", err)
	}
	if err := manager.tasks["runtime-owned"].run(); !errors.Is(err, ErrStopped) {
		t.Fatalf("cron admission ignored parent cancellation: %v", err)
	}
	if err := lifecycle.Stop(t.Context()); err != nil {
		t.Fatal(err)
	}
	if invocations.Load() != 1 {
		t.Fatalf("unexpected callback count %d", invocations.Load())
	}
}
