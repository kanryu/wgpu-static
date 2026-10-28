package wgpu

// UseStaticSymbols is retained for source compatibility. Static calls are
// resolved by the Hike/SXSO-generated internal/wgpunative package.
func UseStaticSymbols(map[string]uintptr) {}
