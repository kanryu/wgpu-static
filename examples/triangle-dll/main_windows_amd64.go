//go:build windows && amd64

package main

import (
	"fmt"
	"runtime"
	"syscall"
	"time"
	"unsafe"

	"github.com/gogpu/gputypes"
)

// This is a Pure Go diagnostic path. It deliberately calls the native DLL
// directly so the DLL boundary can be inspected independently of the Hike
// static-link trampolines.
var (
	nativeDLL                     = syscall.NewLazyDLL("wgpu_native.dll")
	setLogLevel                   = nativeDLL.NewProc("wgpuSetLogLevel")
	getVersion                    = nativeDLL.NewProc("wgpuGetVersion")
	createInstance                = nativeDLL.NewProc("wgpuCreateInstance")
	instanceRelease               = nativeDLL.NewProc("wgpuInstanceRelease")
	instanceCreateSurface         = nativeDLL.NewProc("wgpuInstanceCreateSurface")
	instanceProcessEvents         = nativeDLL.NewProc("wgpuInstanceProcessEvents")
	surfaceRelease                = nativeDLL.NewProc("wgpuSurfaceRelease")
	instanceRequestAdapter        = nativeDLL.NewProc("wgpuInstanceRequestAdapter")
	adapterRelease                = nativeDLL.NewProc("wgpuAdapterRelease")
	adapterRequestDevice          = nativeDLL.NewProc("wgpuAdapterRequestDevice")
	deviceRelease                 = nativeDLL.NewProc("wgpuDeviceRelease")
	deviceGetQueue                = nativeDLL.NewProc("wgpuDeviceGetQueue")
	queueRelease                  = nativeDLL.NewProc("wgpuQueueRelease")
	deviceCreateShaderModule      = nativeDLL.NewProc("wgpuDeviceCreateShaderModule")
	shaderModuleRelease           = nativeDLL.NewProc("wgpuShaderModuleRelease")
	deviceCreateRenderPipeline    = nativeDLL.NewProc("wgpuDeviceCreateRenderPipeline")
	renderPipelineRelease         = nativeDLL.NewProc("wgpuRenderPipelineRelease")
	surfaceGetCurrentTexture      = nativeDLL.NewProc("wgpuSurfaceGetCurrentTexture")
	textureCreateView             = nativeDLL.NewProc("wgpuTextureCreateView")
	textureViewRelease            = nativeDLL.NewProc("wgpuTextureViewRelease")
	textureRelease                = nativeDLL.NewProc("wgpuTextureRelease")
	deviceCreateCommandEncoder    = nativeDLL.NewProc("wgpuDeviceCreateCommandEncoder")
	commandEncoderBeginRenderPass = nativeDLL.NewProc("wgpuCommandEncoderBeginRenderPass")
	renderPassEncoderSetPipeline  = nativeDLL.NewProc("wgpuRenderPassEncoderSetPipeline")
	renderPassEncoderDraw         = nativeDLL.NewProc("wgpuRenderPassEncoderDraw")
	renderPassEncoderEnd          = nativeDLL.NewProc("wgpuRenderPassEncoderEnd")
	renderPassEncoderRelease      = nativeDLL.NewProc("wgpuRenderPassEncoderRelease")
	commandEncoderFinish          = nativeDLL.NewProc("wgpuCommandEncoderFinish")
	commandEncoderRelease         = nativeDLL.NewProc("wgpuCommandEncoderRelease")
	commandBufferRelease          = nativeDLL.NewProc("wgpuCommandBufferRelease")
	queueSubmit                   = nativeDLL.NewProc("wgpuQueueSubmit")
	surfacePresent                = nativeDLL.NewProc("wgpuSurfacePresent")
	surfaceGetCapabilities        = nativeDLL.NewProc("wgpuSurfaceGetCapabilities")
	surfaceFreeCapabilities       = nativeDLL.NewProc("wgpuSurfaceCapabilitiesFreeMembers")
	surfaceConfigure              = nativeDLL.NewProc("wgpuSurfaceConfigure")
)

type dllChainedStruct struct {
	Next  uintptr
	SType uint32
	Pad   uint32
}

type dllSurfaceSourceWindowsHWND struct {
	Chain     dllChainedStruct
	HInstance uintptr
	HWND      uintptr
}

type dllStringView struct {
	Data   uintptr
	Length uintptr
}

type dllSurfaceDescriptor struct {
	Next  uintptr
	Label dllStringView
}

type dllSurfaceCapabilities struct {
	NextInChain      uintptr
	Usages           uint64
	FormatCount      uintptr
	Formats          uintptr
	PresentModeCount uintptr
	PresentModes     uintptr
	AlphaModeCount   uintptr
	AlphaModes       uintptr
}

type dllSurfaceConfiguration struct {
	NextInChain     uintptr
	Device          uintptr
	Format          uint32
	Pad             uint32
	Usage           uint64
	Width           uint32
	Height          uint32
	ViewFormatCount uintptr
	ViewFormats     uintptr
	AlphaMode       uint32
	PresentMode     uint32
}

type dllShaderSourceWGSL struct {
	Chain dllChainedStruct
	Code  dllStringView
}

type dllShaderModuleDescriptor struct {
	Next  uintptr
	Label dllStringView
}

type dllVertexState struct {
	NextInChain   uintptr
	Module        uintptr
	EntryPoint    dllStringView
	ConstantCount uintptr
	Constants     uintptr
	BufferCount   uintptr
	Buffers       uintptr
}

type dllPrimitiveState struct {
	NextInChain      uintptr
	Topology         uint32
	StripIndexFormat uint32
	FrontFace        uint32
	CullMode         uint32
	UnclippedDepth   uint32
	Pad              uint32
}

type dllMultisampleState struct {
	NextInChain     uintptr
	Count           uint32
	Mask            uint32
	AlphaToCoverage uint32
	Pad             uint32
}

type dllColorTargetState struct {
	NextInChain uintptr
	Format      uint32
	Pad         uint32
	Blend       uintptr
	WriteMask   uint64
}

type dllFragmentState struct {
	NextInChain   uintptr
	Module        uintptr
	EntryPoint    dllStringView
	ConstantCount uintptr
	Constants     uintptr
	TargetCount   uintptr
	Targets       uintptr
}

type dllRenderPipelineDescriptor struct {
	NextInChain  uintptr
	Label        dllStringView
	Layout       uintptr
	Vertex       dllVertexState
	Primitive    dllPrimitiveState
	DepthStencil uintptr
	Multisample  dllMultisampleState
	Fragment     uintptr
}

type dllSurfaceTexture struct {
	NextInChain uintptr
	Texture     uintptr
	Status      uint32
	Pad         uint32
}

type dllTextureViewDescriptor struct {
	NextInChain     uintptr
	Label           dllStringView
	Format          uint32
	Dimension       uint32
	BaseMipLevel    uint32
	MipLevelCount   uint32
	BaseArrayLayer  uint32
	ArrayLayerCount uint32
	Aspect          uint32
	Pad             uint32
	Usage           uint64
}

type dllCommandEncoderDescriptor struct {
	Next  uintptr
	Label dllStringView
}

type dllRenderPassColorAttachment struct {
	NextInChain   uintptr
	View          uintptr
	DepthSlice    uint32
	Pad           uint32
	ResolveTarget uintptr
	LoadOp        uint32
	StoreOp       uint32
	ClearR        float64
	ClearG        float64
	ClearB        float64
	ClearA        float64
}

type dllRenderPassDescriptor struct {
	NextInChain          uintptr
	Label                dllStringView
	ColorAttachmentCount uintptr
	ColorAttachments     uintptr
	DepthStencil         uintptr
	OcclusionQuerySet    uintptr
	TimestampWrites      uintptr
}

type dllCommandBufferDescriptor struct {
	Next  uintptr
	Label dllStringView
}

type dllRequestAdapterOptions struct {
	NextInChain       uintptr
	FeatureLevel      uint32
	PowerPreference   uint32
	ForceFallback     uint32
	BackendType       uint32
	CompatibleSurface uintptr
}

type dllCallbackInfo struct {
	NextInChain uintptr
	Mode        uint32
	Pad         uint32
	Callback    uintptr
	Userdata1   uintptr
	Userdata2   uintptr
}

type dllRequestState struct {
	Status      uintptr
	Handle      uintptr
	MessageData uintptr
	MessageLen  uintptr
	Done        uintptr
}

var (
	dllAdapterCallback = syscall.NewCallback(func(status, adapter, messageView, userdata1, userdata2 uintptr) uintptr {
		state := (*dllRequestState)(unsafe.Pointer(userdata1))
		message := (*dllStringView)(unsafe.Pointer(messageView))
		state.Status, state.Handle, state.MessageData, state.MessageLen, state.Done = status, adapter, message.Data, message.Length, 1
		return 0
	})
	dllDeviceCallback = syscall.NewCallback(func(status, device, messageView, userdata1, userdata2 uintptr) uintptr {
		state := (*dllRequestState)(unsafe.Pointer(userdata1))
		message := (*dllStringView)(unsafe.Pointer(messageView))
		state.Status, state.Handle, state.MessageData, state.MessageLen, state.Done = status, device, message.Data, message.Length, 1
		return 0
	})
)

func waitDLLRequest(instance, future uintptr, state *dllRequestState) uintptr {
	for state.Done == 0 {
		instanceProcessEvents.Call(instance)
		time.Sleep(time.Millisecond)
	}
	return state.Handle
}

func main() {
	runtime.LockOSThread()

	if err := nativeDLL.Load(); err != nil {
		panic(fmt.Errorf("load wgpu_native.dll: %w", err))
	}
	window, err := newDLLWindow()
	if err != nil {
		panic(fmt.Errorf("create window: %w", err))
	}
	defer window.close()

	// WGPULogLevel_Off. Keeping this call explicit makes the native call
	// sequence easy to compare with the static-link sample in a debugger.
	if _, _, err := setLogLevel.Call(0); err != syscall.Errno(0) {
		fmt.Printf("wgpuSetLogLevel: %v\n", err)
	}

	version, _, err := getVersion.Call()
	fmt.Printf("wgpu-native DLL version: %d (err=%v)\n", version, err)

	instance, _, err := createInstance.Call(0)
	if instance == 0 {
		panic(fmt.Errorf("wgpuCreateInstance returned NULL: %w", err))
	}
	fmt.Printf("wgpuCreateInstance returned: 0x%x\n", instance)
	defer instanceRelease.Call(instance)

	source := dllSurfaceSourceWindowsHWND{
		Chain:     dllChainedStruct{SType: 5},
		HInstance: window.hinstance,
		HWND:      window.hwnd,
	}
	descriptor := dllSurfaceDescriptor{Next: uintptr(unsafe.Pointer(&source))}
	surface, _, err := instanceCreateSurface.Call(instance, uintptr(unsafe.Pointer(&descriptor)))
	if surface == 0 {
		panic(fmt.Errorf("wgpuInstanceCreateSurface returned NULL: %w", err))
	}
	fmt.Printf("wgpuInstanceCreateSurface returned: 0x%x\n", surface)
	defer surfaceRelease.Call(surface)

	adapterOptions := dllRequestAdapterOptions{CompatibleSurface: surface}
	adapterState := new(dllRequestState)
	adapterCallbackInfo := dllCallbackInfo{Mode: 2, Callback: dllAdapterCallback, Userdata1: uintptr(unsafe.Pointer(adapterState))}
	future, _, err := instanceRequestAdapter.Call(instance, uintptr(unsafe.Pointer(&adapterOptions)), uintptr(unsafe.Pointer(&adapterCallbackInfo)))
	adapter := waitDLLRequest(instance, future, adapterState)
	if adapter == 0 {
		panic(fmt.Errorf("wgpuInstanceRequestAdapter failed: status=%d err=%w", adapterState.Status, err))
	}
	fmt.Printf("wgpuInstanceRequestAdapter returned: 0x%x\n", adapter)
	defer adapterRelease.Call(adapter)

	deviceState := new(dllRequestState)
	deviceCallbackInfo := dllCallbackInfo{Mode: 2, Callback: dllDeviceCallback, Userdata1: uintptr(unsafe.Pointer(deviceState))}
	future, _, err = adapterRequestDevice.Call(adapter, 0, uintptr(unsafe.Pointer(&deviceCallbackInfo)))
	device := waitDLLRequest(instance, future, deviceState)
	if device == 0 {
		panic(fmt.Errorf("wgpuAdapterRequestDevice failed: status=%d err=%w", deviceState.Status, err))
	}
	fmt.Printf("wgpuAdapterRequestDevice returned: 0x%x\n", device)
	defer deviceRelease.Call(device)

	queue, _, err := deviceGetQueue.Call(device)
	if queue == 0 {
		panic(fmt.Errorf("wgpuDeviceGetQueue returned NULL: %w", err))
	}
	fmt.Printf("wgpuDeviceGetQueue returned: 0x%x\n", queue)
	defer queueRelease.Call(queue)

	var capabilities dllSurfaceCapabilities
	result, _, capErr := surfaceGetCapabilities.Call(surface, adapter, uintptr(unsafe.Pointer(&capabilities)))
	if result == 0 || capabilities.FormatCount == 0 || capabilities.Formats == 0 {
		panic(fmt.Errorf("wgpuSurfaceGetCapabilities failed: result=%d err=%w", result, capErr))
	}
	formats := unsafe.Slice((*uint32)(unsafe.Pointer(capabilities.Formats)), capabilities.FormatCount)
	alphaModes := unsafe.Slice((*uint32)(unsafe.Pointer(capabilities.AlphaModes)), capabilities.AlphaModeCount)
	presentModes := unsafe.Slice((*uint32)(unsafe.Pointer(capabilities.PresentModes)), capabilities.PresentModeCount)
	fmt.Printf("surface capabilities: formats=%d present_modes=%d alpha_modes=%d\n", len(formats), len(presentModes), len(alphaModes))
	alphaMode := uint32(gputypes.CompositeAlphaModeAuto)
	if len(alphaModes) > 0 {
		alphaMode = alphaModes[0]
	}
	presentMode := uint32(gputypes.PresentModeFifo)
	if len(presentModes) > 0 {
		presentMode = presentModes[0]
	}
	config := dllSurfaceConfiguration{
		Device: device, Format: formats[0], Usage: uint64(gputypes.TextureUsageRenderAttachment),
		Width: 640, Height: 480, AlphaMode: alphaMode, PresentMode: presentMode,
	}
	surfaceConfigure.Call(surface, uintptr(unsafe.Pointer(&config)))
	surfaceFreeCapabilities.Call(uintptr(unsafe.Pointer(&capabilities)))
	fmt.Printf("surface configured: format=%d\n", formats[0])

	shaderCode := []byte("@vertex\nfn vs_main(@builtin(vertex_index) i: u32) -> @builtin(position) vec4f {\n  var p = array<vec2f, 3>(vec2f(0.0, 0.5), vec2f(-0.5, -0.5), vec2f(0.5, -0.5));\n  return vec4f(p[i], 0.0, 1.0);\n}\n@fragment\nfn fs_main() -> @location(0) vec4f { return vec4f(0.0, 0.8, 0.2, 1.0); }\n")
	shaderSource := dllShaderSourceWGSL{
		Chain: dllChainedStruct{SType: 2},
		Code:  dllStringView{Data: uintptr(unsafe.Pointer(&shaderCode[0])), Length: uintptr(len(shaderCode))},
	}
	shaderDescriptor := dllShaderModuleDescriptor{Next: uintptr(unsafe.Pointer(&shaderSource))}
	shader, _, shaderErr := deviceCreateShaderModule.Call(device, uintptr(unsafe.Pointer(&shaderDescriptor)))
	if shader == 0 {
		panic(fmt.Errorf("wgpuDeviceCreateShaderModule returned NULL: %w", shaderErr))
	}
	fmt.Printf("wgpuDeviceCreateShaderModule returned: 0x%x\n", shader)
	defer shaderModuleRelease.Call(shader)
	runtime.KeepAlive(shaderCode)

	vertexEntry := []byte("vs_main")
	fragmentEntry := []byte("fs_main")
	colorTarget := dllColorTargetState{Format: formats[0], WriteMask: 0xF}
	fragment := dllFragmentState{
		Module:      shader,
		EntryPoint:  dllStringView{Data: uintptr(unsafe.Pointer(&fragmentEntry[0])), Length: uintptr(len(fragmentEntry))},
		TargetCount: 1,
		Targets:     uintptr(unsafe.Pointer(&colorTarget)),
	}
	pipelineDescriptor := dllRenderPipelineDescriptor{
		Vertex: dllVertexState{
			Module:     shader,
			EntryPoint: dllStringView{Data: uintptr(unsafe.Pointer(&vertexEntry[0])), Length: uintptr(len(vertexEntry))},
		},
		Primitive:   dllPrimitiveState{},
		Multisample: dllMultisampleState{Count: 1, Mask: 0xFFFFFFFF},
		Fragment:    uintptr(unsafe.Pointer(&fragment)),
	}
	pipeline, _, pipelineErr := deviceCreateRenderPipeline.Call(device, uintptr(unsafe.Pointer(&pipelineDescriptor)))
	if pipeline == 0 {
		panic(fmt.Errorf("wgpuDeviceCreateRenderPipeline returned NULL: %w", pipelineErr))
	}
	fmt.Printf("wgpuDeviceCreateRenderPipeline returned: 0x%x\n", pipeline)
	defer renderPipelineRelease.Call(pipeline)
	runtime.KeepAlive(vertexEntry)
	runtime.KeepAlive(fragmentEntry)
	runtime.KeepAlive(colorTarget)
	runtime.KeepAlive(fragment)
	runtime.KeepAlive(pipelineDescriptor)

	frames := 0
	for window.poll() {
		var current dllSurfaceTexture
		surfaceGetCurrentTexture.Call(surface, uintptr(unsafe.Pointer(&current)))
		if current.Status != 1 && current.Status != 2 {
			instanceProcessEvents.Call(instance)
			continue
		}
		view, _, viewErr := textureCreateView.Call(current.Texture, 0)
		if view == 0 {
			panic(fmt.Errorf("wgpuTextureCreateView returned NULL: %w", viewErr))
		}
		colorAttachment := dllRenderPassColorAttachment{
			View: view, DepthSlice: 0xFFFFFFFF, LoadOp: 2, StoreOp: 1,
			ClearR: 0.08, ClearG: 0.08, ClearB: 0.12, ClearA: 1.0,
		}
		passDescriptor := dllRenderPassDescriptor{
			ColorAttachmentCount: 1,
			ColorAttachments:     uintptr(unsafe.Pointer(&colorAttachment)),
		}
		encoderDescriptor := dllCommandEncoderDescriptor{}
		encoder, _, encoderErr := deviceCreateCommandEncoder.Call(device, uintptr(unsafe.Pointer(&encoderDescriptor)))
		if encoder == 0 {
			panic(fmt.Errorf("wgpuDeviceCreateCommandEncoder returned NULL: %w", encoderErr))
		}
		pass, _, passErr := commandEncoderBeginRenderPass.Call(encoder, uintptr(unsafe.Pointer(&passDescriptor)))
		if pass == 0 {
			panic(fmt.Errorf("wgpuCommandEncoderBeginRenderPass returned NULL: %w", passErr))
		}
		renderPassEncoderSetPipeline.Call(pass, pipeline)
		renderPassEncoderDraw.Call(pass, 3, 1, 0, 0)
		renderPassEncoderEnd.Call(pass)
		renderPassEncoderRelease.Call(pass)
		commandDescriptor := dllCommandBufferDescriptor{}
		commandBuffer, _, finishErr := commandEncoderFinish.Call(encoder, uintptr(unsafe.Pointer(&commandDescriptor)))
		commandEncoderRelease.Call(encoder)
		if commandBuffer == 0 {
			panic(fmt.Errorf("wgpuCommandEncoderFinish returned NULL: %w", finishErr))
		}
		queueSubmit.Call(queue, 1, uintptr(unsafe.Pointer(&commandBuffer)))
		commandBufferRelease.Call(commandBuffer)
		surfacePresent.Call(surface)
		textureViewRelease.Call(view)
		textureRelease.Call(current.Texture)
		frames++
		if frames == 1 || frames%120 == 0 {
			fmt.Printf("rendered frames: %d\n", frames)
		}
		instanceProcessEvents.Call(instance)
	}
	fmt.Printf("render loop stopped after %d frames\n", frames)
}
