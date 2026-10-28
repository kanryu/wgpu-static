//go:build windows

package wgpu

import (
	"sync"
	"syscall"
)

var callbackCompileMu sync.Mutex

// newCallback creates a native callback entry point without the GoFFI loader.
// The callback entry functions use the Windows ABI expected by wgpu-native.
func newCallback(fn any) uintptr {
	callbackCompileMu.Lock()
	defer callbackCompileMu.Unlock()
	return syscall.NewCallback(fn)
}
