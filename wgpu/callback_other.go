//go:build !windows

package wgpu

// Non-Windows static targets currently require a platform-specific callback
// trampoline supplied by the Hike/SXSO layer.
func newCallback(fn any) uintptr { return 0 }
