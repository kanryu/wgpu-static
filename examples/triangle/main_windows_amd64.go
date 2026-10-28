//go:build windows && amd64

package main

import (
	"fmt"
	"os"
	"runtime"
	"syscall"
	"unsafe"

	"github.com/gogpu/gputypes"
	"github.com/kanryu/wgpu-static/internal/wgpunative"
	_ "github.com/kanryu/wgpu-static/internal/wgpunative_native"
	"github.com/kanryu/wgpu-static/wgpu"
)

const (
	windowWidth                           = 640
	windowHeight                          = 480
	wmDestroy                             = 0x0002
	pmRemove                              = 0x0001
	wmQuit                                = 0x0012
	csHRedraw                             = 0x0002
	csVRedraw                             = 0x0001
	swShow                                = 5
	wsOverlapped                          = 0x00CF0000
	wsVisible                             = 0x10000000
	errorClassAlreadyExists syscall.Errno = 1410
)

type wndClassEx struct {
	cbSize, style          uint32
	wndProc                uintptr
	clsExtra, wndExtra     int32
	hInstance, hIcon       uintptr
	hCursor, hbrBackground uintptr
	menuName, className    *uint16
	hIconSm                uintptr
}

type point struct{ x, y int32 }
type msg struct {
	hwnd, message, wParam, lParam, time uintptr
	pt                                  point
}
type rect struct{ left, top, right, bottom int32 }

var (
	user32             = syscall.NewLazyDLL("user32.dll")
	kernel32           = syscall.NewLazyDLL("kernel32.dll")
	registerClassExW   = user32.NewProc("RegisterClassExW")
	createWindowExW    = user32.NewProc("CreateWindowExW")
	defWindowProcW     = user32.NewProc("DefWindowProcW")
	destroyWindow      = user32.NewProc("DestroyWindow")
	dispatchMessageW   = user32.NewProc("DispatchMessageW")
	getClientRect      = user32.NewProc("GetClientRect")
	windowProcCallback uintptr
	peekMessageW       = user32.NewProc("PeekMessageW")
	postQuitMessage    = user32.NewProc("PostQuitMessage")
	showWindow         = user32.NewProc("ShowWindow")
	translateMessage   = user32.NewProc("TranslateMessage")
	getModuleHandleW   = kernel32.NewProc("GetModuleHandleW")
)

func windowProc(hwnd, message, wParam, lParam uintptr) uintptr {
	if message == wmDestroy {
		postQuitMessage.Call(0)
		return 0
	}
	ret, _, _ := defWindowProcW.Call(hwnd, message, wParam, lParam)
	return ret
}

type nativeWindow struct {
	hwnd      uintptr
	hinstance uintptr
}

func newWindow(title string) (*nativeWindow, error) {
	className, _ := syscall.UTF16PtrFromString("WGPUStaticTriangle")
	titleName, _ := syscall.UTF16PtrFromString(title)
	instance, _, _ := getModuleHandleW.Call(0)
	windowProcCallback = syscall.NewCallback(windowProc)
	class := wndClassEx{
		cbSize: uint32(unsafe.Sizeof(wndClassEx{})), style: csHRedraw | csVRedraw,
		wndProc: windowProcCallback, hInstance: instance,
		className: className,
	}
	if atom, _, err := registerClassExW.Call(uintptr(unsafe.Pointer(&class))); atom == 0 && err != errorClassAlreadyExists {
		return nil, fmt.Errorf("RegisterClassExW: %w", err)
	}
	hwnd, _, err := createWindowExW.Call(0, uintptr(unsafe.Pointer(className)), uintptr(unsafe.Pointer(titleName)),
		wsOverlapped|wsVisible, 100, 100, windowWidth, windowHeight, 0, 0, instance, 0)
	if hwnd == 0 {
		return nil, fmt.Errorf("CreateWindowExW: %w", err)
	}
	showWindow.Call(hwnd, swShow)
	return &nativeWindow{hwnd: hwnd, hinstance: instance}, nil
}

func (w *nativeWindow) size() (uint32, uint32) {
	var r rect
	getClientRect.Call(w.hwnd, uintptr(unsafe.Pointer(&r)))
	return uint32(r.right - r.left), uint32(r.bottom - r.top)
}

func (w *nativeWindow) poll() bool {
	var m msg
	// WM_QUIT is thread-wide and has no HWND. Use a zero HWND filter so the
	// main loop receives it after WM_DESTROY posts it.
	for processed := 0; processed < 64; processed++ {
		result, _, _ := peekMessageW.Call(uintptr(unsafe.Pointer(&m)), 0, 0, 0, pmRemove)
		if result == 0 {
			break
		}
		if m.message == wmQuit {
			return false
		}
		translateMessage.Call(uintptr(unsafe.Pointer(&m)))
		dispatchMessageW.Call(uintptr(unsafe.Pointer(&m)))
	}
	return true
}

func (w *nativeWindow) close() { destroyWindow.Call(w.hwnd) }

// requestAdapterWhilePumping keeps the Win32 thread responsive while
// wgpu-native completes its asynchronous adapter request.
func requestAdapterWhilePumping(window *nativeWindow, instance *wgpu.Instance, options *wgpu.RequestAdapterOptions) (*wgpu.Adapter, error) {
	return instance.RequestAdapterWithEventPump(options, window.poll)
}

// probeRequestAdapterABI is an opt-in ABI probe. It intentionally bypasses
// wgpu.Instance.RequestAdapter so the generated InstanceRequestAdapterABI
// entry can be stopped in a debugger with recognizable argument bytes.
func probeRequestAdapterABI(instance *wgpu.Instance, surface *wgpu.Surface) {
	if os.Getenv("WGPU_ABI_PROBE") != "1" {
		return
	}

	type probeAdapterOptions struct {
		NextInChain          uintptr
		FeatureLevel         uint32
		PowerPreference      uint32
		ForceFallbackAdapter uint32
		BackendType          uint32
		CompatibleSurface    uintptr
	}

	state := new(wgpunative.RequestAdapterCallbackState)
	options := probeAdapterOptions{
		NextInChain:          0x1111111111111111,
		FeatureLevel:         0x22222222,
		PowerPreference:      0x33333333,
		ForceFallbackAdapter: 0x44444444,
		BackendType:          0x55555555,
		CompatibleSurface:    0x6666666666666666,
	}
	info := wgpunative.CallbackInfoWire{
		NextInChain: 0x7777777777777777,
		Mode:        0x8888888888888888,
		Callback:    0x9999999999999999,
		Userdata1:   0xAAAAAAAAAAAAAAAA,
		Userdata2:   0xBBBBBBBBBBBBBBBB,
	}

	instanceHandle := uintptr(unsafe.Pointer(instance.Handle()))
	if surface != nil {
		options.CompatibleSurface = uintptr(unsafe.Pointer(surface.Handle()))
	}
	if os.Getenv("WGPU_ABI_PROBE_SAFE") == "1" {
		options = probeAdapterOptions{CompatibleSurface: options.CompatibleSurface}
		info = wgpunative.CallbackInfoWire{
			Mode:      2, // WGPUCallbackModeAllowProcessEvents
			Callback:  wgpunative.RequestAdapterCallbackPtr(),
			Userdata1: uintptr(unsafe.Pointer(state)),
		}
		fmt.Println("ABI probe: using valid WGPU enum values and callback pointer")
	}
	fmt.Printf("ABI probe instance=%#x options=%#x info=%#x\n", instanceHandle, uintptr(unsafe.Pointer(&options)), uintptr(unsafe.Pointer(&info)))
	fmt.Printf("ABI probe options=%+x info=%+x state=%#x\n", options, info, uintptr(unsafe.Pointer(state)))
	fmt.Println("ABI probe: calling InstanceRequestAdapterABI")
	future := wgpunative.InstanceRequestAdapterABI(
		(*byte)(unsafe.Pointer(instanceHandle)),
		(*byte)(unsafe.Pointer(&options)),
		info,
	)
	fmt.Printf("ABI probe returned future=%#x state=%+x\n", future, *state)
}

// requestDeviceWhilePumping is the device equivalent of
// requestAdapterWhilePumping.
func requestDeviceWhilePumping(window *nativeWindow, adapter *wgpu.Adapter) (*wgpu.Device, error) {
	return adapter.RequestDeviceWithEventPump(nil, window.poll)
}

func main() {
	runtime.LockOSThread()
	fmt.Println("initializing wgpu-native")
	if err := wgpu.Init(); err != nil {
		panic(err)
	}
	wgpunative.SetLogLevel(0) // WGPULogLevel_Off: avoid debug-output notifications.
	fmt.Printf("wgpu-native version: %d\n", wgpunative.GetVersion())
	// Register all Go callbacks before creating the Win32 window. Creating the
	// window can start dispatching messages on the OS thread while callback
	// trampolines are still being registered.
	wgpu.InitializeCallbacks()

	window, err := newWindow("triangle [wgpu-native + Go]")
	if err != nil {
		panic(err)
	}
	defer window.close()

	fmt.Println("creating instance")
	instance, err := wgpu.CreateInstance(nil)
	if err != nil {
		panic(err)
	}
	defer instance.Release()
	surface, err := instance.CreateSurfaceFromWindowsHWND(window.hinstance, window.hwnd)
	if err != nil {
		panic(err)
	}
	defer surface.Release()
	probeRequestAdapterABI(instance, surface)
	// The debug static build currently corrupts the Windows syscall state when
	// printing immediately after the native surface call; keep this checkpoint
	// silent while isolating the FFI trampoline.
	// Request the adapter without embedding the HWND-backed surface in the
	// asynchronous request. The surface is validated immediately afterwards by
	// GetCapabilities; this avoids blocking adapter selection on Win32 surface
	// compatibility callbacks.
	adapter, err := requestAdapterWhilePumping(window, instance, &wgpu.RequestAdapterOptions{CompatibleSurface: surface})
	if err != nil {
		panic(err)
	}
	defer adapter.Release()
	fmt.Println("requesting device")
	device, err := requestDeviceWhilePumping(window, adapter)
	if err != nil {
		panic(err)
	}
	defer device.Release()
	queue := device.Queue()
	defer queue.Release()
	fmt.Println("creating pipeline")

	shader, err := device.CreateShaderModuleWGSL(shaderSource)
	if err != nil {
		panic(err)
	}
	defer shader.Release()
	layout, err := device.CreatePipelineLayoutSimple(nil)
	if err != nil {
		panic(err)
	}
	defer layout.Release()
	caps, err := surface.GetCapabilities(adapter)
	if err != nil || len(caps.Formats) == 0 {
		panic(fmt.Errorf("surface capabilities: %w", err))
	}
	width, height := window.size()
	if width == 0 || height == 0 {
		width, height = windowWidth, windowHeight
	}
	if err := surface.Configure(device, &wgpu.SurfaceConfiguration{
		Format: caps.Formats[0], Usage: gputypes.TextureUsageRenderAttachment,
		Width: width, Height: height, AlphaMode: caps.AlphaModes[0], PresentMode: gputypes.PresentModeFifo,
	}); err != nil {
		panic(err)
	}
	pipeline, err := device.CreateRenderPipelineSimple(layout, shader, "vs_main", shader, "fs_main", caps.Formats[0])
	if err != nil {
		panic(err)
	}
	defer pipeline.Release()

	// Main Win32 event/render loop. Keep pumping messages even when there is
	// no pending message; rendering is performed once per iteration.
	running := true
	for running {
		running = window.poll()
		if !running {
			break
		}
		frame, _, err := surface.GetCurrentTexture()
		if err != nil {
			continue
		}
		view, err := frame.Texture.CreateView(nil)
		if err != nil {
			panic(err)
		}
		encoder, err := device.CreateCommandEncoder(nil)
		if err != nil {
			panic(err)
		}
		pass, err := encoder.BeginRenderPass(&wgpu.RenderPassDescriptor{ColorAttachments: []wgpu.RenderPassColorAttachment{{
			View: view, LoadOp: gputypes.LoadOpClear, StoreOp: gputypes.StoreOpStore,
			ClearValue: wgpu.Color{R: 0.0, G: 0.2, B: 0.0, A: 1.0},
		}}})
		if err != nil {
			panic(err)
		}
		pass.SetPipeline(pipeline)
		pass.Draw(3, 1, 0, 0)
		pass.End()
		commands, err := encoder.Finish()
		if err != nil {
			panic(err)
		}
		_, _ = queue.Submit(commands)
		surface.Present(frame)
		commands.Release()
		encoder.Release()
		view.Release()
		frame.Texture.Release()
	}
}

const shaderSource = `@vertex
fn vs_main(@builtin(vertex_index) index: u32) -> @builtin(position) vec4<f32> {
    let x = f32(i32(index) - 1);
    let y = f32(i32(index & 1u) * 2 - 1);
    return vec4<f32>(x, y, 0.0, 1.0);
}
@fragment
fn fs_main() -> @location(0) vec4<f32> {
    return vec4<f32>(1.0, 0.0, 0.0, 1.0);
}`
