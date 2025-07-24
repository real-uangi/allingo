/*
 * Copyright 2025 Uangi. All rights reserved.
 * @author uangi
 * @date 2025/4/9 12:46
 */

// Package ready

package ready

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"
)

func (m *Manager) HandleHealth(w http.ResponseWriter) {
	buffer := bytes.NewBuffer(make([]byte, 0, 1024))
	health := true
	err := m.ctx.ForEach(func(name string, cp CheckPoint) error {
		start := time.Now()
		err := cp.Ready()
		if err != nil {
			health = false
			_, _ = fmt.Fprintf(buffer, "%s FAILED %v\n", name, err)
		} else {
			_, _ = fmt.Fprintf(buffer, "%s OK %.3fs\n", name, time.Since(start).Seconds())
		}
		return nil
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = fmt.Fprintf(buffer, "Manager FAILED %v\n", err)
		return
	}
	if health {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
	}
	_, _ = io.Copy(w, buffer)
}

func (m *Manager) HandleHealthTarget(w http.ResponseWriter, target string) {
	if target == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	cp, ok := m.ctx.Get(target)
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	start := time.Now()
	err := cp.Ready()
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = fmt.Fprintf(w, "%s FAILED %v\n", target, err)
	} else {
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprintf(w, "%s OK %.3fs\n", target, time.Since(start).Seconds())
	}
}
