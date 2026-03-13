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
	"time"
)

var logger = log.NewStdLogger("async")

func execLongRun(f Function) func() {
	return func() {
		panicked := true
		var err error
		defer func() {
			if err != nil {
				logger.Errorf(err.(error), "long-run func returns error: %s, stack:\n %s", err.(error).Error(), trace.Stack(3))
			}
			if panicked {
				err := recover()
				if err != nil {
					logger.Errorf(err.(error), "long-run func panic: %s, stack:\n %s", err.(error).Error(), trace.Stack(3))
				}
			}
		}()
		err = f()
		panicked = false
	}
}

func seal(f Function, second int, keepRunning bool) func() {
	superId := goid.Get()
	if keepRunning {
		if second != 0 {
			return func() {
				defer holder.Clear()
				holder.Put(constants.TraceIdHeader, trace.GetSpecific(superId))
				time.Sleep(time.Duration(second) * time.Second)
				for {
					execLongRun(f)()
					logger.Warn("exited long-run goroutine, restart in 5 seconds")
					time.Sleep(5 * time.Second)
				}
			}
		}
		return func() {
			defer holder.Clear()
			holder.Put(constants.TraceIdHeader, trace.GetSpecific(superId))
			for {
				execLongRun(f)()
				logger.Warn("exited long-run goroutine, restart in 5 seconds")
				time.Sleep(5 * time.Second)
			}
		}
	} else {
		if second != 0 {
			return func() {
				panicked := true
				var err error
				defer func() {
					if err != nil {
						logger.Errorf(err.(error), "func returns error: %s, stack:\n %s", err.(error).Error(), trace.Stack(3))
					}
					if panicked {
						err := recover()
						if err != nil {
							logger.Errorf(err.(error), "func panic: %s, stack:\n %s", err.(error).Error(), trace.Stack(3))
						}
					}
				}()
				defer holder.Clear()
				time.Sleep(time.Duration(second) * time.Second)
				holder.Put(constants.TraceIdHeader, trace.GetSpecific(superId))
				err = f()
				panicked = false
			}
		}
		return func() {
			panicked := true
			var err error
			defer func() {
				if err != nil {
					logger.Errorf(err.(error), "func returns error: %s, stack:\n %s", err.(error).Error(), trace.Stack(3))
				}
				if panicked {
					err := recover()
					if err != nil {
						logger.Errorf(err.(error), "func panic: %s, stack:\n %s", err.(error).Error(), trace.Stack(3))
					}
				}
			}()
			defer holder.Clear()
			holder.Put(constants.TraceIdHeader, trace.GetSpecific(superId))
			err = f()
			panicked = false
		}
	}

}

// submit seals Function and handover it to the pool
func submit(f Function, delaySeconds int, keepRunning bool) error {
	sealed := seal(f, 0, keepRunning)
	if keepRunning {
		go sealed()
		return nil
	} else {
		return pool.Submit(sealed)
	}
}

func SubmitOnce(f Function) error {
	return submit(f, 0, false)
}

func SubmitKeepRunning(f Function) error {
	return submit(f, 0, true)
}

func SubmitDelayOnce(f Function, seconds int) error {
	return submit(f, seconds, false)
}

func SubmitDelayKeepRunning(f Function, seconds int) error {
	return submit(f, seconds, true)
}

func ExitTimeout(second int) {
	err := pool.ReleaseTimeout(time.Duration(second) * time.Second)
	if err != nil {
		logger.Errorf(err, "exit timeout")
	}
}
