/*
 * Copyright 2025 Uangi. All rights reserved.
 * @author uangi
 * @date 2025/8/3 15:35
 */

// Package performance

package performance

import (
	"github.com/prometheus/client_golang/prometheus"
)

var activeSSEGauge = prometheus.NewGauge(prometheus.GaugeOpts{
	Name: "active_sse_connections",
	Help: "Number of currently active SSE connections",
})

func init() {
	_ = prometheus.Register(activeSSEGauge)
}

func SSEBegin() {
	activeSSEGauge.Inc()
}

func SSEEnd() {
	activeSSEGauge.Dec()
}
