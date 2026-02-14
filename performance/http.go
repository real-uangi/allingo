/*
 * Copyright 2025 Uangi. All rights reserved.
 * @author uangi
 * @date 2025/8/3 15:59
 */

// Package performance

package performance

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"strconv"
	"strings"
	"time"
)

const (
	ms float64 = 0.001
	s  float64 = 1
)

var (
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "请求总数统计",
		},
		[]string{"method", "path", "status"},
	)

	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "请求持续时间统计（秒）",
			Buckets: []float64{50 * ms, 200 * ms, 500 * ms, s, 5 * s},
		},
		[]string{"method", "path"},
	)
)

func init() {
	prometheus.MustRegister(httpRequestsTotal, httpRequestDuration)
}

func GinHttpMiddleware(c *gin.Context) {
	path := c.FullPath()
	if !strings.HasPrefix(path, "/api/") {
		c.Next()
		return
	}

	start := time.Now()
	c.Next()
	duration := time.Since(start).Seconds()
	status := c.Writer.Status()
	httpRequestsTotal.WithLabelValues(c.Request.Method, path, strconv.Itoa(status)).Inc()
	httpRequestDuration.WithLabelValues(c.Request.Method, path).Observe(duration)
}
