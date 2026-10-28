package wgpu

// InitializeCallbacks is retained for API compatibility. Adapter and device
// request callbacks now use native Hike/SXSO C-ABI entries and therefore do
// not need syscall.NewCallback registration here.
func InitializeCallbacks() {
}
