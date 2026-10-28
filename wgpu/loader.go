package wgpu

// These types remain only for source compatibility with the unexported
// declarations left in older generated code. All live WGPU calls use
// internal/wgpunative directly; there is no dynamic loader implementation.
type Library interface {
	NewProc(string) Proc
}

type Proc interface {
	Call(args ...uintptr) (uintptr, uintptr, error)
}

// Library represents a dynamically loaded library (DLL/SO/DYLIB).
// Platform-specific implementations handle the actual loading mechanism.

// Proc represents a procedure (function pointer) from a dynamically loaded library.
// It abstracts platform-specific function calling mechanisms.

// float32Proc is implemented by platform loaders for procedures whose native
// return type is float32. Proc.Call intentionally keeps the existing integer
// return contract for the rest of the WebGPU API; this narrow interface lets
// those procedures use the platform's floating-point return ABI instead.
