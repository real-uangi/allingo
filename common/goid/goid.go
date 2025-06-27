//go:build (386 || amd64 || amd64p32 || arm || arm64) && gc

package goid

// Defined in goid_go1.5.s.
func getg() *g

func Get() int64 {
	return getg().goid
}
