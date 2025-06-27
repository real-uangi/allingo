/*
 * Copyright © 2025 Uangi. All rights reserved.
 * @author uangi
 * @date 2025/2/25 16:31
 */

// Package debug
package debug

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"
)

func RunPprofHttpServer(port string) {
	_ = http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
}
