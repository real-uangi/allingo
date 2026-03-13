/*
 * Copyright 2026 Uangi. All rights reserved.
 * @author uangi
 * @date 2026/3/13 09:38
 */

// Package async

package async_test

import (
	"github.com/real-uangi/allingo/common/async"
	"sync"
	"testing"
	"time"
)

func TestAsync(t *testing.T) {
	const batchSize = 1000
	var wg sync.WaitGroup
	start := time.Now()
	for i := 0; i < batchSize; i++ {
		wg.Add(1)
		err := async.SubmitOnce(func() error {
			defer wg.Done()
			t.Logf("task %d at %.3fs", i, time.Since(start).Seconds())
			return nil
		})
		if err != nil {
			wg.Done()
			t.Log(err)
		}
	}
	wg.Wait()
}
