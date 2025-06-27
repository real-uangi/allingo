/*
 * Copyright 2025 Uangi. All rights reserved.
 * @author uangi
 * @date 2025/6/16 17:04
 */

// Package log

package log

import (
	"go.uber.org/fx/fxevent"
)

type FxLogger struct {
	logger *StdLogger
}

func NewFxLogger() fxevent.Logger {
	return &FxLogger{
		logger: For[fxevent.Logger](),
	}
}

func (fl *FxLogger) LogEvent(event fxevent.Event) {
	switch e := event.(type) {
	case *fxevent.OnStartExecuting:
		break
	case *fxevent.OnStartExecuted:
		if e.Err == nil {
			fl.logger.Infof("started [%s], took %dms", e.FunctionName, e.Runtime.Milliseconds())
		} else {
			fl.logger.Errorf(e.Err, "start [%s]", e.FunctionName)
		}
		break
	case *fxevent.OnStopExecuting:
		break
	case *fxevent.OnStopExecuted:
		if e.Err == nil {
			fl.logger.Infof("stopped [%s], took %dms", e.FunctionName, e.Runtime.Milliseconds())
		} else {
			fl.logger.Errorf(e.Err, "stopp [%s]", e.FunctionName)
		}
		break
	case *fxevent.Supplied:
		fl.logger.Infof("Supplied: Type=%s, Err=%v", e.TypeName, e.Err)
		break
	case *fxevent.Provided:
		if e.Err == nil {
			fl.logger.Infof("provided %v from %s", e.OutputTypeNames, e.ConstructorName)
		} else {
			fl.logger.Errorf(e.Err, "failed to privide %v from %s", e.ConstructorName, e.OutputTypeNames)
		}
		break
	case *fxevent.Replaced:
		fl.logger.Warnf("replaced %v", e.OutputTypeNames)
		break
	case *fxevent.Decorated:
		fl.logger.Infof("decorated %v to %s", e.OutputTypeNames, e.DecoratorName)
		break
	case *fxevent.Run:
		if e.Err == nil {
			fl.logger.Infof("succeed %s [%s] %dms", e.Kind, e.Name, e.Runtime.Milliseconds())
		} else {
			fl.logger.Errorf(e.Err, "failed %s [%s]", e.Kind, e.Name)
		}
		break
	case *fxevent.Invoking:
		break
	case *fxevent.Invoked:
		if e.Err == nil {
			fl.logger.Infof("invoked [%s]", e.FunctionName)
		} else {
			fl.logger.Errorf(e.Err, "failed to invoke [%s]", e.FunctionName)
		}
		break
	case *fxevent.Started:
		if e.Err == nil {
			fl.logger.Infof("server started")
		}
		break
	case *fxevent.Stopping:
		fl.logger.Infof("server stopping: signal=%v", e.Signal)
		break
	case *fxevent.Stopped:
		if e.Err == nil {
			fl.logger.Infof("server stopped")
		} else {
			fl.logger.Errorf(e.Err, "failed to stop server")
		}
		break
	case *fxevent.RollingBack:
		fl.logger.Warnf("rollingBack from startError %v", e.StartErr)
		break
	case *fxevent.RolledBack:
		if e.Err == nil {
			fl.logger.Warnf("rolled back")
		} else {
			fl.logger.Errorf(e.Err, "failed to rolled back")
		}
		break
	case *fxevent.BeforeRun:
		break
	case *fxevent.LoggerInitialized:
		break
	default:
		fl.logger.Warnf("Unknown event type: %T", e)
	}
}
