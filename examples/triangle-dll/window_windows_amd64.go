//go:build windows && amd64

package main

import (
	"syscall"
	"unsafe"
)

const (
	wmDestroy    = 0x0002
	wmQuit       = 0x0012
	pmRemove     = 0x0001
	csHRedraw    = 0x0002
	csVRedraw    = 0x0001
	wsOverlapped = 0x00CF0000
	wsVisible    = 0x10000000
	swShow       = 5
)

type dllWndClass struct {
	cbSize, style          uint32
	wndProc                uintptr
	clsExtra, wndExtra     int32
	hInstance, hIcon       uintptr
	hCursor, hbrBackground uintptr
	menuName, className    *uint16
	hIconSm                uintptr
}

type dllMsg struct {
	hwnd, message, wParam, lParam, time uintptr
	pt                                  struct{ x, y int32 }
}

var (
	dllUser32           = syscall.NewLazyDLL("user32.dll")
	dllKernel32         = syscall.NewLazyDLL("kernel32.dll")
	dllRegisterClassEx  = dllUser32.NewProc("RegisterClassExW")
	dllCreateWindowEx   = dllUser32.NewProc("CreateWindowExW")
	dllDefWindowProc    = dllUser32.NewProc("DefWindowProcW")
	dllDestroyWindow    = dllUser32.NewProc("DestroyWindow")
	dllDispatchMessage  = dllUser32.NewProc("DispatchMessageW")
	dllPeekMessage      = dllUser32.NewProc("PeekMessageW")
	dllPostQuitMessage  = dllUser32.NewProc("PostQuitMessage")
	dllShowWindow       = dllUser32.NewProc("ShowWindow")
	dllTranslateMessage = dllUser32.NewProc("TranslateMessage")
	dllGetModuleHandle  = dllKernel32.NewProc("GetModuleHandleW")
)

func dllWindowProc(hwnd, message, wParam, lParam uintptr) uintptr {
	if message == wmDestroy {
		dllPostQuitMessage.Call(0)
		return 0
	}
	ret, _, _ := dllDefWindowProc.Call(hwnd, message, wParam, lParam)
	return ret
}

type dllWindow struct {
	hwnd, hinstance uintptr
}

func newDLLWindow() (*dllWindow, error) {
	className, _ := syscall.UTF16PtrFromString("WGPUStaticDLLTriangle")
	title, _ := syscall.UTF16PtrFromString("wgpu-native DLL triangle")
	instance, _, _ := dllGetModuleHandle.Call(0)
	proc := syscall.NewCallback(dllWindowProc)
	class := dllWndClass{cbSize: uint32(unsafe.Sizeof(dllWndClass{})), style: csHRedraw | csVRedraw, wndProc: proc, hInstance: instance, className: className}
	dllRegisterClassEx.Call(uintptr(unsafe.Pointer(&class)))
	hwnd, _, err := dllCreateWindowEx.Call(0, uintptr(unsafe.Pointer(className)), uintptr(unsafe.Pointer(title)), wsOverlapped|wsVisible, 100, 100, 640, 480, 0, 0, instance, 0)
	if hwnd == 0 {
		return nil, err
	}
	dllShowWindow.Call(hwnd, swShow)
	return &dllWindow{hwnd: hwnd, hinstance: instance}, nil
}

func (w *dllWindow) poll() bool {
	var message dllMsg
	for i := 0; i < 64; i++ {
		ok, _, _ := dllPeekMessage.Call(uintptr(unsafe.Pointer(&message)), 0, 0, 0, pmRemove)
		if ok == 0 {
			return true
		}
		if message.message == wmQuit {
			return false
		}
		dllTranslateMessage.Call(uintptr(unsafe.Pointer(&message)))
		dllDispatchMessage.Call(uintptr(unsafe.Pointer(&message)))
	}
	return true
}

func (w *dllWindow) close() { dllDestroyWindow.Call(w.hwnd) }
