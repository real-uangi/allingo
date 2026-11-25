/*
 * Copyright © 2024 Uangi. All rights reserved.
 * @author uangi
 * @date 2024/11/28 10:38
 */

// Package task
package task

import (
	"context"
	"errors"
	"strconv"
	"sync"
	"time"

	"github.com/real-uangi/allingo/common/async"
	"github.com/real-uangi/allingo/common/log"
	"github.com/real-uangi/allingo/common/trace"
	"github.com/real-uangi/allingo/kv"
	"github.com/robfig/cron/v3"
	"go.uber.org/fx"
)

const taskCounterTTL = 365 * 24 * time.Hour

var ErrStopped = errors.New("task manager is stopped")
var ErrNotStarted = errors.New("task manager is not started")
var ErrTaskNotFound = errors.New("task is not registered")

type TaskManager struct {
	c       *cron.Cron
	logger  *log.StdLogger
	tasks   map[string]*Task
	mu      sync.Mutex
	app     string
	kv      kv.KV
	ctx     context.Context
	cancel  context.CancelFunc
	running map[string]bool
	work    sync.WaitGroup
	stopped bool
	started bool
	done    chan struct{}
}

func NewManager(kv kv.KV) *TaskManager {
	manager := newManager(kv, context.Background())
	manager.c.Start()
	manager.started = true
	return manager
}

// NewManagedManager starts scheduling only after Fx startup and drains the
// actual async work on shutdown. Its running context is independent of the
// short-lived context supplied to an OnStart hook.
func NewManagedManager(lifecycle fx.Lifecycle, store kv.KV) *TaskManager {
	return NewManagedManagerWithContext(lifecycle, store, context.Background())
}

// NewManagedManagerWithContext also withdraws admission and cancels running
// callbacks when the application's runtime ends, before a later OnStop hook.
func NewManagedManagerWithContext(lifecycle fx.Lifecycle, store kv.KV, parent context.Context) *TaskManager {
	manager := newManager(store, parent)
	lifecycle.Append(fx.Hook{
		OnStart: func(context.Context) error {
			manager.mu.Lock()
			defer manager.mu.Unlock()
			if manager.stopped || manager.ctx.Err() != nil {
				return ErrStopped
			}
			manager.c.Start()
			manager.started = true
			return nil
		},
		OnStop: manager.Stop,
	})
	return manager
}

func newManager(store kv.KV, parent context.Context) *TaskManager {
	ctx, cancel := context.WithCancel(parent)
	manager := &TaskManager{
		c: cron.New(cron.WithParser(cron.NewParser(
			cron.SecondOptional | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor,
		))),
		logger:  log.For[TaskManager](),
		tasks:   make(map[string]*Task),
		kv:      store,
		ctx:     ctx,
		cancel:  cancel,
		running: make(map[string]bool),
		done:    make(chan struct{}),
	}
	return manager
}

type Task struct {
	taskId  string
	id      cron.EntryID
	counter int64
	kv      kv.KV
	logger  *log.StdLogger
	run     func() error
}

// Add cron "秒 分 时 d日 m月 w周几 @...ly"
func (manager *TaskManager) Add(name, spec string, f func() error) (*Task, error) {
	return manager.AddContext(name, spec, func(context.Context) error { return f() })
}

// AddContext registers a task whose context is canceled when Stop begins.
// A task name has at most one queued or running invocation in this manager.
func (manager *TaskManager) AddContext(name, spec string, f func(context.Context) error) (*Task, error) {
	manager.mu.Lock()
	defer manager.mu.Unlock()
	if manager.stopped || manager.ctx.Err() != nil {
		return nil, ErrStopped
	}
	manager.logger.Infof("add task %s with spec: %s", name, spec)
	if t := manager.tasks[name]; t != nil {
		manager.logger.Warnf("task already existed: %s", name)
	}
	var counter int64 = 0
	var err error
	taskId := "COCK_CLIENT:TASK:" + name
	last, ok, err := manager.kv.Get(taskId)
	if err != nil {
		return nil, err
	}
	if ok && last != "" {
		counter, err = strconv.ParseInt(last, 10, 64)
		if err != nil {
			return nil, err
		}
	}
	t := &Task{
		taskId:  taskId,
		counter: counter,
		kv:      manager.kv,
		logger:  log.NewStdLogger("allingo.task." + name),
	}
	t.run = func() error {
		// Admission and Stop share a lock, so no new WaitGroup work can be
		// added after shutdown starts waiting for accepted async submissions.
		manager.mu.Lock()
		if manager.stopped || manager.ctx.Err() != nil {
			manager.mu.Unlock()
			return ErrStopped
		}
		if !manager.started {
			manager.mu.Unlock()
			return ErrNotStarted
		}
		if manager.running[name] {
			manager.mu.Unlock()
			return nil
		}
		manager.running[name] = true
		manager.work.Add(1)
		manager.mu.Unlock()
		finish := func() {
			manager.mu.Lock()
			delete(manager.running, name)
			manager.mu.Unlock()
			manager.work.Done()
		}
		err := async.SubmitOnce(func() error {
			defer finish()
			if manager.ctx.Err() != nil {
				return nil
			}
			var err error
			//验证
			if !t.occupy() {
				manager.logger.Infof("ignoring task [%s]", name)
				return nil
			}
			if manager.ctx.Err() != nil {
				return nil
			}
			manager.logger.Infof("task [%s] will be executed on this machine", name)
			panicked := true
			start := time.Now()
			defer func() {
				// 记录任务执行日志
				execTime := time.Now().UnixMilli() - start.UnixMilli()
				if panicked {
					ev := recover()
					if ev != nil {
						manager.logger.Errorf(err, "task [%s] panic: %v, stack:\n %s", name, ev, trace.Stack(3))
					}
				} else if err != nil {
					manager.logger.Errorf(err, "task [%s] return error %v", name, err)
				} else {
					manager.logger.Infof("task [%s] successfully executed, cost %d ms", name, execTime)
				}
			}()
			// just call user function with warp
			err = f(manager.ctx)
			panicked = false
			return nil
		})
		if err != nil {
			finish()
			manager.logger.Errorf(err, "failed to submit task [%s]", name)
		}
		return err
	}
	id, err := manager.c.AddFunc(spec, func() { _ = t.run() })
	if err != nil {
		return nil, err
	}
	t.id = id
	manager.tasks[name] = t
	return t, nil
}

// Trigger requests one immediate invocation through the same admission,
// non-overlap and async-drain path used by cron. An already active invocation
// is not queued again.
func (manager *TaskManager) Trigger(name string) error {
	manager.mu.Lock()
	if manager.stopped || manager.ctx.Err() != nil {
		manager.mu.Unlock()
		return ErrStopped
	}
	if !manager.started {
		manager.mu.Unlock()
		return ErrNotStarted
	}
	t := manager.tasks[name]
	manager.mu.Unlock()
	if t == nil {
		return ErrTaskNotFound
	}
	return t.run()
}

// MustAdd cron "秒 分 时 d日 m月 w周几 @...ly"
func (manager *TaskManager) MustAdd(name, spec string, f func() error) {
	_, err := manager.Add(name, spec, f)
	if err != nil {
		panic(err)
	}
}

func (manager *TaskManager) MustAddContext(name, spec string, f func(context.Context) error) {
	if _, err := manager.AddContext(name, spec, f); err != nil {
		panic(err)
	}
}

// Stop stops scheduling, cancels task contexts and waits for queued and running
// async work. A deadline only bounds the caller's wait; callbacks which ignore
// cancellation continue running, and a later Stop can wait for them again.
func (manager *TaskManager) Stop(ctx context.Context) error {
	manager.mu.Lock()
	if !manager.stopped {
		manager.stopped = true
		manager.cancel()
		cronStopped := manager.c.Stop()
		go func() {
			<-cronStopped.Done()
			manager.work.Wait()
			close(manager.done)
		}()
	}
	manager.mu.Unlock()
	select {
	case <-manager.done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (task *Task) occupy() bool {
	if ok, err := task.tryOccupy(); err != nil {
		task.logger.Error(err)
		return false
	} else {
		return ok
	}
}

func (task *Task) tryOccupy() (bool, error) {
	if task.counter == 0 {
		ok, err := task.kv.SetIfAbsent(task.taskId, "1", taskCounterTTL)
		if err != nil {
			return false, err
		}
		if ok {
			task.counter = 1
			return true, nil
		}
	}

	latestCounter, latestValue, ok, err := task.loadLatestCounter()
	if err != nil {
		return false, err
	}
	if !ok || latestCounter == 0 {
		ok, err := task.kv.SetIfAbsent(task.taskId, "1", taskCounterTTL)
		if err != nil {
			return false, err
		}
		if ok {
			task.counter = 1
			return true, nil
		}
		latestCounter, _, ok, err = task.loadLatestCounter()
		if err != nil {
			return false, err
		}
		if ok {
			task.counter = latestCounter
		}
		return false, nil
	}
	if latestCounter != task.counter {
		task.counter = latestCounter
		return false, nil
	}

	nextCounter := latestCounter + 1
	ok, err = task.kv.CompareAndSet(task.taskId, latestValue, strconv.FormatInt(nextCounter, 10), taskCounterTTL)
	if err != nil {
		return false, err
	}
	if ok {
		task.counter = nextCounter
		return true, nil
	}

	latestCounter, _, ok, err = task.loadLatestCounter()
	if err != nil {
		return false, err
	}
	if ok {
		task.counter = latestCounter
	}
	return false, nil
}

func (task *Task) loadLatestCounter() (int64, string, bool, error) {
	last, ok, err := task.kv.Get(task.taskId)
	if err != nil {
		return 0, "", false, err
	}
	if !ok || last == "" {
		return 0, "", false, nil
	}
	counter, err := strconv.ParseInt(last, 10, 64)
	if err != nil {
		return 0, last, false, err
	}
	return counter, last, true, nil
}
