/*
 * Copyright 2026 Uangi. All rights reserved.
 * @author uangi
 * @date 2026/3/13 09:46
 */

// Package async

package async

import "github.com/prometheus/client_golang/prometheus"

var (
	capacityGauge = prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name: "async_pool_capacity",
			Help: "Capacity of this pool.",
		},
		func() float64 {
			return float64(pool.Cap())
		},
	)
	runningGauge = prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name: "async_pool_running",
			Help: "Number of workers currently running.",
		},
		func() float64 {
			return float64(pool.Running())
		},
	)
	waitingGauge = prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name: "async_pool_waiting",
			Help: "Number of tasks waiting to be executed.",
		},
		func() float64 {
			return float64(pool.Waiting())
		},
	)
)

func init() {
	prometheus.MustRegister(capacityGauge, runningGauge, waitingGauge)
}
