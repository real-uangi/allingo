/*
 * Copyright © 2024 Uangi. All rights reserved.
 * @author uangi
 * @date 2024/11/28 10:38
 */

// Package task
package task

import (
	"github.com/real-uangi/allingo/common/async"
	"github.com/real-uangi/allingo/common/log"
	"github.com/real-uangi/allingo/common/trace"
	"github.com/real-uangi/allingo/kv"
	"github.com/robfig/cron/v3"
	"strconv"
	"sync"
	"time"
)

const taskCounterTTL = 365 * 24 * time.Hour

type TaskManager struct {
	c      *cron.Cron
	logger *log.StdLogger
	tasks  map[string]*Task
	mu     sync.Mutex
	app    string
	kv     kv.KV
}

func NewManager(kv kv.KV) *TaskManager {
	manager := &TaskManager{
		c: cron.New(cron.WithParser(cron.NewParser(
			cron.SecondOptional | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor,
		))),
		logger: log.For[TaskManager](),
		tasks:  make(map[string]*Task),
		kv:     kv,
	}
	manager.c.Start()
	return manager
}

type Task struct {
	taskId  string
	id      cron.EntryID
	counter int64
	kv      kv.KV
	logger  *log.StdLogger
}

// Add cron "秒 分 时 d日 m月 w周几 @...ly"
func (manager *TaskManager) Add(name, spec string, f func() error) (*Task, error) {
	manager.mu.Lock()
	defer manager.mu.Unlock()
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
	id, err := manager.c.AddFunc(spec, func() {
		//执行
		_ = async.SubmitOnce(func() error {
			var err error
			//验证
			if !t.occupy() {
				manager.logger.Infof("ignoring task [%s]", name)
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
			err = f()
			panicked = false
			return nil
		})
	})
	if err != nil {
		return nil, err
	}
	t.id = id
	manager.tasks[name] = t
	return t, nil
}

// MustAdd cron "秒 分 时 d日 m月 w周几 @...ly"
func (manager *TaskManager) MustAdd(name, spec string, f func() error) {
	_, err := manager.Add(name, spec, f)
	if err != nil {
		panic(err)
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
