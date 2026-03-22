/*
 * Copyright 2025 Uangi. All rights reserved.
 * @author uangi
 * @date 2025/6/16 17:04
 */

// Package log

package log

import (
	"go.uber.org/fx/fxevent"
	"time"
)

type FxLogger struct {
	logger *StdLogger
	initAt time.Time
}

func NewFxLogger() fxevent.Logger {
	return &FxLogger{
		logger: For[fxevent.Logger](),
		initAt: time.Now(),
	}
}

func (fl *FxLogger) LogEvent(event fxevent.Event) {
	switch e := event.(type) {
	case *fxevent.OnStartExecuting:
		fl.logger.Infof("Starting [%s] by [%s]", e.FunctionName, e.CallerName)
		break
	case *fxevent.OnStartExecuted:
		if e.Err == nil {
			fl.logger.Infof("Succeed starting [%s], took %dms", e.FunctionName, e.Runtime.Milliseconds())
		} else {
			fl.logger.Errorf(e.Err, "Failed starting [%s]", e.FunctionName)
		}
		break
	case *fxevent.OnStopExecuting:
		fl.logger.Infof("Stopping [%s] by [%s]", e.FunctionName, e.CallerName)
		break
	case *fxevent.OnStopExecuted:
		if e.Err == nil {
			fl.logger.Infof("Stopped [%s], took %dms", e.FunctionName, e.Runtime.Milliseconds())
		} else {
			fl.logger.Errorf(e.Err, "Failed stopping [%s]", e.FunctionName)
		}
		break
	case *fxevent.Supplied:
		if e.Err == nil {
			fl.logger.Infof("Supplied type[%s/%s]", e.ModuleName, e.TypeName)
		} else {
			fl.logger.Errorf(e.Err, "Failed suppling [%s/%s]\n%v", e.ModuleName, e.TypeName, e.StackTrace)
		}
		break
	case *fxevent.Provided:
		if e.Err == nil {
			fl.logger.Infof("Module [%s] provided [%s] from %s", e.ModuleName, e.OutputTypeNames, e.ConstructorName)
		} else {
			fl.logger.Errorf(e.Err, "Module [%s] failed to privide %v from %s", e.ModuleName, e.ConstructorName, e.OutputTypeNames)
		}
		break
	case *fxevent.Replaced:
		fl.logger.Warnf("Replaced %s/%v", e.ModuleName, e.OutputTypeNames)
		break
	case *fxevent.Decorated:
		fl.logger.Infof("Decorated %v to %s in module %s", e.OutputTypeNames, e.DecoratorName, e.ModuleName)
		break
	case *fxevent.BeforeRun:
		fl.logger.Infof("Doing %s [%s] in module [%s]", e.Kind, e.Name, e.ModuleName)
		break
	case *fxevent.Run:
		if e.Err == nil {
			fl.logger.Infof("Succeed %s [%s] in module [%s] %.3fs", e.Kind, e.Name, e.ModuleName, e.Runtime.Seconds())
		} else {
			fl.logger.Errorf(e.Err, "Failed to %s [%s]", e.Kind, e.Name)
		}
		break
	case *fxevent.Invoking:
		fl.logger.Infof("Invoking [%s/%s]", e.ModuleName, e.FunctionName)
		break
	case *fxevent.Invoked:
		if e.Err != nil {
			fl.logger.Errorf(e.Err, "Failed to invoke [%s/%s]\n%s", e.ModuleName, e.FunctionName, e.Trace)
		}
		break
	case *fxevent.Started:
		if e.Err == nil {
			fl.logger.Infof("Server initlization completed after %.3f seconds", time.Since(fl.initAt).Seconds())
		}
		break
	case *fxevent.Stopping:
		fl.logger.Infof("Server stopping: signal=%v", e.Signal)
		break
	case *fxevent.Stopped:
		if e.Err == nil {
			fl.logger.Infof("Server stopped. Total running time %s", time.Since(fl.initAt).String())
		} else {
			fl.logger.Errorf(e.Err, "Failed to stop server")
		}
		break
	case *fxevent.RollingBack:
		fl.logger.Warnf("RollingBack from startError %v", e.StartErr)
		break
	case *fxevent.RolledBack:
		if e.Err == nil {
			fl.logger.Warnf("Rolled back")
		} else {
			fl.logger.Errorf(e.Err, "Failed to rolled back")
		}
		break
	case *fxevent.LoggerInitialized:
		break
	default:
		fl.logger.Warnf("Unknown event type: %T", e)
	}
}
