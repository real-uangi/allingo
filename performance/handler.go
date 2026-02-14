/*
 * Copyright 2025 Uangi. All rights reserved.
 * @author uangi
 * @date 2025/8/3 15:33
 */

// Package performance

package performance

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func init() {
	_ = prometheus.Register(collectors.NewGoCollector())
	_ = prometheus.Register(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
}

func Handler() gin.HandlerFunc {
	return gin.WrapH(promhttp.Handler())
}
