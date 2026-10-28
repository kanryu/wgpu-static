package wgpu

import "unsafe"

func nativeBytePtr(value uintptr) *byte {
	if value == 0 {
		return nil
	}
	return (*byte)(unsafe.Pointer(value))
}

func nativeHandle(value *byte) uintptr {
	return uintptr(unsafe.Pointer(value))
}
