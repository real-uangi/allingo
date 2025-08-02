/*
 * Copyright © 2024 Uangi. All rights reserved.
 * @author uangi
 * @date 2024/11/25 15:43
 */

// Package async
package async

import (
	"github.com/real-uangi/allingo/common/constants"
	"github.com/real-uangi/allingo/common/goid"
	"github.com/real-uangi/allingo/common/holder"
	"github.com/real-uangi/allingo/common/log"
	"github.com/real-uangi/allingo/common/trace"
	"sync"
	"time"
)

var asyncWaitGroup sync.WaitGroup
var logger = log.NewStdLogger("async")

func execLongRun(f func(), second int) func() {
	return func() {
		panicked := true
		defer func() {
			if panicked {
				err := recover()
				if err != nil {
					logger.Errorf(err.(error), "async long-run func panic: %s, stack:\n %s", err.(error).Error(), trace.Stack(3))
				}
			}
		}()
		time.Sleep(time.Duration(second) * time.Second)
		f()
		panicked = false
	}
}

func seal(f func(), second int, longRun bool) func() {
	superId := goid.Get()
	if longRun {
		if second != 0 {
			return func() {
				defer holder.Clear()
				holder.Put(constants.TraceIdHeader, trace.GetSpecific(superId))
				time.Sleep(time.Duration(second) * time.Second)
				for {
					execLongRun(f, second)()
					logger.Warn("exited long run thread, restart in 5 seconds")
					time.Sleep(5 * time.Second)
				}
			}
		}
		return func() {
			defer holder.Clear()
			holder.Put(constants.TraceIdHeader, trace.GetSpecific(superId))
			for {
				execLongRun(f, second)()
				logger.Warn("exited long run thread, restart in 5 seconds")
				time.Sleep(5 * time.Second)
			}
		}
	} else {
		if second != 0 {
			return func() {
				asyncWaitGroup.Add(1)
				defer asyncWaitGroup.Done()
				panicked := true
				defer func() {
					if panicked {
						err := recover()
						if err != nil {
							logger.Errorf(err.(error), "async func panic: %s, stack:\n %s", err.(error).Error(), trace.Stack(3))
						}
					}
				}()
				defer holder.Clear()
				time.Sleep(time.Duration(second) * time.Second)
				holder.Put(constants.TraceIdHeader, trace.GetSpecific(superId))
				f()
				panicked = false
			}
		}
		return func() {
			asyncWaitGroup.Add(1)
			defer asyncWaitGroup.Done()
			panicked := true
			defer func() {
				if panicked {
					err := recover()
					if err != nil {
						logger.Errorf(err.(error), "async func panic: %s, stack:\n %s", err.(error).Error(), trace.Stack(3))
					}
				}
			}()
			defer holder.Clear()
			holder.Put(constants.TraceIdHeader, trace.GetSpecific(superId))
			f()
			panicked = false
		}
	}

}

// Go execute the job immediately
func Go(f func(), longRun bool) {
	go seal(f, 0, longRun)()
}

func DoOnce(f func()) {
	Go(f, false)
}

func DoHold(f func()) {
	Go(f, true)
}

func Delay(f func(), seconds int, longRun bool) {
	go seal(f, seconds, longRun)()
}

func DelayOnce(f func(), seconds int) {
	Delay(f, seconds, false)
}

func DelayHold(f func(), seconds int) {
	Delay(f, seconds, true)
}

func ExitTimeout(second int) {
	select {
	case <-time.After(time.Duration(second) * time.Second):
		logger.Warn("shutdown timeout!")
		return
	case <-func() chan int {
		c := make(chan int)
		go func() {
			asyncWaitGroup.Wait()
			c <- 0
		}()
		return c
	}():
		return
	}
}
