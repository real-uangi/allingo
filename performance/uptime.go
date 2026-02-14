/*
 * Copyright 2025 Uangi. All rights reserved.
 * @author uangi
 * @date 2025/4/7 09:51
 */

// Package performance
package performance

import (
	"github.com/prometheus/client_golang/prometheus"
	"time"
)

func init() {
	// 记录应用启动时间
	startTime := time.Now()

	// 创建一个 GaugeFunc，用于记录运行时长（单位：秒）
	uptimeGauge := prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name: "app_uptime_seconds",
			Help: "Application uptime in seconds.",
		},
		func() float64 {
			return time.Since(startTime).Seconds()
		},
	)

	// 注册这个指标
	prometheus.MustRegister(uptimeGauge)
}
