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
	"slices"
	"strconv"
	"sync"
	"time"
)

type TaskManager struct {
	c      *cron.Cron
	logger *log.StdLogger
	added  []string
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
		added:  make([]string, 0),
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
	lock    kv.Lock
	logger  *log.StdLogger
}

// Add cron "秒 分 时 d日 m月 w周几 @...ly"
func (manager *TaskManager) Add(name, spec string, f func() error) (*Task, error) {
	manager.logger.Infof("add task %s with spec: %s", name, spec)
	if !manager.checkName(name) {
		manager.logger.Warnf("task name already existed: %s", name)
	}
	var counter int64 = 0
	var err error
	taskId := "COCK_CLIENT:TASK:" + name
	last, ok := manager.kv.Get(taskId)
	if ok && last != "" {
		counter, err = strconv.ParseInt(last, 10, 64)
		if err != nil {
			return nil, err
		}
	}
	t := &Task{
		taskId:  taskId,
		counter: counter,
		lock:    manager.kv.NewLock(taskId + ":lock"),
		kv:      manager.kv,
		logger:  log.NewStdLogger("marmot.dTask." + name),
	}
	id, err := manager.c.AddFunc(spec, func() {
		//执行
		async.Go(func() {
			var err error
			//验证
			if !t.occupy() {
				manager.logger.Infof("ignoring task [%s]", name)
				return
			}
			manager.logger.Infof("task [%s] will be executed", name)
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
		}, false)
	})
	if err != nil {
		return nil, err
	}
	t.id = id
	return t, nil
}

func (manager *TaskManager) checkName(name string) bool {
	manager.mu.Lock()
	defer manager.mu.Unlock()
	result := !slices.Contains(manager.added, name)
	manager.added = append(manager.added, name)
	return result
}

func (task *Task) occupy() bool {
	err := task.lock.Lock(24*time.Hour, time.Second)
	if err != nil {
		task.logger.Error(err)
		return false
	}
	defer task.lock.Unlock()
	var latestCounter int64 = 0
	defer func() {
		task.counter = latestCounter
	}()
	last, _ := task.kv.Get(task.taskId)
	if last != "" {
		latestCounter, err = strconv.ParseInt(last, 10, 64)
	}
	if latestCounter == 0 || latestCounter == task.counter {
		latestCounter++
		err := task.kv.SetStruct(task.taskId, strconv.FormatInt(latestCounter, 10), 365*24*time.Hour)
		if err != nil {
			task.logger.Error(err, "error occurs when occupying task")
		}
		return true
	}
	return false
}
