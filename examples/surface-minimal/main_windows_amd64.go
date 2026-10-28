//go:build windows && amd64

package main

import (
	"fmt"
	"runtime"
	"syscall"
	"unsafe"

	"github.com/kanryu/wgpu-static/internal/wgpunative"
	_ "github.com/kanryu/wgpu-static/internal/wgpunative_native"
)

const (
	className = "WGPUStaticMinimal"
	wsVisible = 0x10000000
	wsWindow  = 0x00CF0000
)

type chainedStruct struct {
	next  uintptr
	sType uint32
	pad   uint32
}

type surfaceSourceWindowsHWND struct {
	chain    chainedStruct
	hinstance uintptr
	hwnd      uintptr
}

type stringView struct {
	data   uintptr
	length uintptr
}

type surfaceDescriptor struct {
	nextInChain uintptr
	label       stringView
}

type wndClassEx struct {
	cbSize, style          uint32
	wndProc                uintptr
	clsExtra, wndExtra     int32
	hInstance, hIcon       uintptr
	hCursor, hbrBackground uintptr
	menuName, className    *uint16
	hIconSm                uintptr
}

var (
	user32          = syscall.NewLazyDLL("user32.dll")
	kernel32        = syscall.NewLazyDLL("kernel32.dll")
	registerClassEx = user32.NewProc("RegisterClassExW")
	createWindowEx  = user32.NewProc("CreateWindowExW")
	defWindowProc   = user32.NewProc("DefWindowProcW")
	destroyWindow   = user32.NewProc("DestroyWindow")
	getModuleHandle = kernel32.NewProc("GetModuleHandleW")
	windowProc      uintptr
)

func wndProc(hwnd, message, wparam, lparam uintptr) uintptr {
	returnValue, _, _ := defWindowProc.Call(hwnd, message, wparam, lparam)
	return returnValue
}

func createWindow(instance uintptr) (uintptr, error) {
	class, _ := syscall.UTF16PtrFromString(className)
	title, _ := syscall.UTF16PtrFromString("wgpu minimal")
	windowProc = syscall.NewCallback(wndProc)
	windowClass := wndClassEx{
		cbSize: uint32(unsafe.Sizeof(wndClassEx{})),
		wndProc: windowProc,
		hInstance: instance,
		className: class,
	}
	if atom, _, err := registerClassEx.Call(uintptr(unsafe.Pointer(&windowClass))); atom == 0 && err != syscall.Errno(1410) {
		return 0, fmt.Errorf("RegisterClassExW: %w", err)
	}
	hwnd, _, err := createWindowEx.Call(0, uintptr(unsafe.Pointer(class)), uintptr(unsafe.Pointer(title)), wsWindow|wsVisible, 100, 100, 640, 480, 0, 0, instance, 0)
	if hwnd == 0 {
		return 0, fmt.Errorf("CreateWindowExW: %w", err)
	}
	return hwnd, nil
}

func main() {
	runtime.LockOSThread()
	instanceHandle, _, _ := getModuleHandle.Call(0)
	hwnd, err := createWindow(instanceHandle)
	if err != nil {
		panic(err)
	}
	defer destroyWindow.Call(hwnd)

	fmt.Printf("wgpuGetVersion: %d\n", wgpunative.GetVersion())
	fmt.Println("setting wgpu log level: trace")
	wgpunative.SetLogLevel(5) // WGPULogLevel_Trace
	fmt.Println("calling wgpuCreateInstance directly")
	instance := wgpunative.CreateInstance(nil)
	if instance == nil {
		panic("wgpuCreateInstance returned NULL")
	}
	fmt.Println("instance created successfully")
	defer wgpunative.InstanceRelease(instance)

	source := surfaceSourceWindowsHWND{
		chain: chainedStruct{sType: 5},
		hinstance: instanceHandle,
		hwnd: hwnd,
	}
	descriptor := surfaceDescriptor{
		nextInChain: uintptr(unsafe.Pointer(&source)),
	}

	fmt.Println("calling wgpuInstanceCreateSurface directly")
	surface := wgpunative.InstanceCreateSurface(instance, (*byte)(unsafe.Pointer(&descriptor)))
	if surface == nil {
		panic("wgpuInstanceCreateSurface returned NULL")
	}
	fmt.Println("surface created successfully")
	defer wgpunative.SurfaceRelease(surface)
}
