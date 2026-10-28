package wgpu

import (
	"sync"
	"unsafe"

	"github.com/gogpu/gputypes"
	"github.com/kanryu/wgpu-static/internal/wgpunative"
)

// RequestDeviceCallbackInfo holds callback configuration for RequestDevice.
type RequestDeviceCallbackInfo struct {
	NextInChain uintptr // *ChainedStruct
	Mode        CallbackMode
	Callback    uintptr // Function pointer
	Userdata1   uintptr
	Userdata2   uintptr
}

// deviceRequest holds state for an async device request.
type deviceRequest struct {
	done    chan struct{}
	device  *Device
	status  RequestDeviceStatus
	message string
}

var (
	// deviceRequests is the global registry for pending device requests.
	// Protected by deviceRequestsMu for concurrent access.
	deviceRequests   = make(map[uintptr]*deviceRequest)
	deviceRequestsMu sync.Mutex
	deviceRequestID  uintptr

	// deviceCallbackPtr is the callback function pointer (created once).
	// Protected by deviceCallbackOnce for concurrent initialization.
	deviceCallbackPtr  uintptr
	deviceCallbackOnce sync.Once
)

// handleDeviceCallback completes a request after the platform callback entry
// normalizes the ABI-specific WGPUStringView representation.
func handleDeviceCallback(status uintptr, device uintptr, message StringView, userdata1 uintptr) uintptr {
	// Find and complete the request
	deviceRequestsMu.Lock()
	req, ok := deviceRequests[userdata1]
	if ok {
		delete(deviceRequests, userdata1)
	}
	deviceRequestsMu.Unlock()

	if ok && req != nil {
		req.status = RequestDeviceStatus(status)
		if device != 0 {
			trackResource(device, "Device")
			req.device = &Device{handle: device}
		}
		req.message = stringViewToString(message)
		close(req.done)
	}
	return 0
}

// initDeviceCallback creates the platform-correct C callback function pointer.
func initDeviceCallback() {
	deviceCallbackPtr = newCallback(deviceCallbackEntry)
}

// RequestDevice requests a GPU device from the adapter.
// This is a synchronous wrapper that blocks until the device is available.
func (a *Adapter) RequestDevice(options *DeviceDescriptor) (*Device, error) {
	return a.requestDevice(options, nil)
}

// RequestDeviceWithEventPump requests a device while keeping callbacks on
// the caller's thread. The adapter's originating instance processes events;
// pump is called after each ProcessEvents call and may return false to cancel.
func (a *Adapter) RequestDeviceWithEventPump(options *DeviceDescriptor, pump func() bool) (*Device, error) {
	return a.requestDevice(options, pump)
}

func (a *Adapter) requestDevice(options *DeviceDescriptor, pump func() bool) (*Device, error) {
	if err := checkInit(); err != nil {
		return nil, err
	}
	if a == nil || a.handle == 0 {
		return nil, &WGPUError{Op: "RequestDevice", Message: "adapter is nil or released"}
	}

	// The native Hike/SXSO callback writes this fixed-layout state block. It
	// does not enter the Go runtime, so no syscall.NewCallback is involved.
	// Keep callback userdata on the heap so native event dispatch cannot retain
	// a Go stack slot while the runtime is handling Windows callbacks.
	callbackState := new(wgpunative.RequestDeviceCallbackState)

	// Convert Go-idiomatic descriptor to wire format.
	var optionsPtr uintptr
	var reqLimitsWire limitsWire // kept alive for the duration of the FFI call
	if options != nil {
		wire := deviceDescriptorWire{
			Label: stringToStringView(options.Label),
		}
		if len(options.RequiredFeatures) > 0 {
			wire.RequiredFeatureCount = uintptr(len(options.RequiredFeatures))
			wire.RequiredFeatures = uintptr(unsafe.Pointer(&options.RequiredFeatures[0]))
		}
		if options.RequiredLimits != nil {
			reqLimitsWire = limitsToWire(options.RequiredLimits)
			wire.RequiredLimits = uintptr(unsafe.Pointer(&reqLimitsWire))
		}
		optionsPtr = uintptr(unsafe.Pointer(&wire))
	}
	_ = reqLimitsWire // ensure not optimised away before the call below

	// Prepare callback info
	callbackInfo := RequestDeviceCallbackInfo{
		NextInChain: 0,
		// Deliver the callback from the originating instance's ProcessEvents
		// call, never from an arbitrary native thread.
		Mode:      CallbackModeAllowProcessEvents,
		Callback:  wgpunative.RequestDeviceCallbackPtr(),
		Userdata1: uintptr(unsafe.Pointer(callbackState)),
		Userdata2: 0,
	}

	// Call wgpuAdapterRequestDevice
	adapterHandle := a.handle
	 wgpunative.AdapterRequestDeviceABI(
		nativeBytePtr(adapterHandle),
		nativeBytePtr(optionsPtr),
		wgpunative.CallbackInfoWire{
			NextInChain: callbackInfo.NextInChain,
			Mode:        uintptr(callbackInfo.Mode),
			Callback:    callbackInfo.Callback,
			Userdata1:   callbackInfo.Userdata1,
			Userdata2:   callbackInfo.Userdata2,
		},
	)

	// Process events until callback fires
	for {
		if callbackState.Done != 0 {
			// Callback completed
			status := RequestDeviceStatus(callbackState.Status)
			if status != RequestDeviceStatusSuccess {
				msg := stringViewToString(StringView{Data: callbackState.MessageData, Length: callbackState.MessageLength})
				if msg == "" {
					msg = "device request failed"
				}
				return nil, &WGPUError{Op: "RequestDevice", Message: msg}
			}
			// Cache limits at creation time so Limits() returns value without FFI.
			if callbackState.Device != 0 {
				device := &Device{handle: callbackState.Device}
				trackResource(callbackState.Device, "Device")
				device.limits = fetchDeviceLimits(device.handle)
				return device, nil
			}
			return nil, &WGPUError{Op: "RequestDevice", Message: "device request returned no device"}
		}

		if a.instance != nil {
			a.instance.ProcessEvents()
		}
		if pump != nil && !pump() {
			return nil, &WGPUError{Op: "RequestDevice", Message: "event pump cancelled"}
		}
	}
}

// fetchDeviceLimits calls wgpuDeviceGetLimits and converts the wire struct to public Limits.
// Returns zero-value Limits on failure (non-fatal: limits remain valid defaults).
func fetchDeviceLimits(handle uintptr) Limits {
	var wire limitsWire
	status := wgpunative.DeviceGetLimits(
		handle,
		uintptr(unsafe.Pointer(&wire)),
	)
	if WGPUStatus(status) != WGPUStatusSuccess {
		return Limits{}
	}
	return limitsFromWire(&wire)
}

// Queue returns the default queue for the device.
func (d *Device) Queue() *Queue {
	mustInit()
	if d == nil || d.handle == 0 {
		return nil
	}
	handle := wgpunative.DeviceGetQueue(nativeBytePtr(d.handle))
	if handle == nil {
		return nil
	}
	native := nativeHandle(handle)
	trackResource(native, "Queue")
	return &Queue{handle: native}
}

// Poll polls the device for completed work.
// If wait is true, blocks until there is work to process.
// Returns true if the queue is empty.
// This is a wgpu-native extension.
func (d *Device) Poll(wait bool) bool {
	mustInit()
	if d == nil || d.handle == 0 {
		return true
	}
	return wgpunative.DevicePoll(nativeBytePtr(d.handle), wait, nil)
}

// Release releases the device resources.
func (d *Device) Release() {
	if d.handle != 0 {
		untrackResource(d.handle)
		wgpunative.DeviceRelease(nativeBytePtr(d.handle))
		d.handle = 0
	}
}

// Release releases the queue resources.
func (q *Queue) Release() {
	if q.handle != 0 {
		untrackResource(q.handle)
		wgpunative.QueueRelease(nativeBytePtr(q.handle))
		q.handle = 0
	}
}

// DeviceLostCallbackInfo configures the device-lost callback.
type DeviceLostCallbackInfo struct {
	NextInChain uintptr // *ChainedStruct
	Mode        CallbackMode
	Callback    uintptr // Function pointer
	Userdata1   uintptr
	Userdata2   uintptr
}

// UncapturedErrorCallbackInfo configures the uncaptured-error callback.
type UncapturedErrorCallbackInfo struct {
	NextInChain uintptr // *ChainedStruct
	Callback    uintptr // Function pointer
	Userdata1   uintptr
	Userdata2   uintptr
}

// DeviceDescriptor configures device creation.
// Matches the gogpu/wgpu API for cross-project compatibility.
type DeviceDescriptor struct {
	// Label is an optional debug label for the device.
	Label string
	// RequiredFeatures lists GPU features that the device must support.
	RequiredFeatures []FeatureName
	// RequiredLimits, if non-nil, specifies minimum resource limits the device must meet.
	// Pass nil to use the adapter's default limits.
	RequiredLimits *Limits
}

// limitsToWire converts public Limits to the FFI-compatible limitsWire struct.
// Used when passing required limits to wgpuAdapterRequestDevice.
func limitsToWire(l *Limits) limitsWire {
	if l == nil {
		return limitsWire{}
	}
	return limitsWire{
		MaxTextureDimension1D:                     l.MaxTextureDimension1D,
		MaxTextureDimension2D:                     l.MaxTextureDimension2D,
		MaxTextureDimension3D:                     l.MaxTextureDimension3D,
		MaxTextureArrayLayers:                     l.MaxTextureArrayLayers,
		MaxBindGroups:                             l.MaxBindGroups,
		MaxBindGroupsPlusVertexBuffers:            l.MaxBindGroupsPlusVertexBuffers,
		MaxBindingsPerBindGroup:                   l.MaxBindingsPerBindGroup,
		MaxDynamicUniformBuffersPerPipelineLayout: l.MaxDynamicUniformBuffersPerPipelineLayout,
		MaxDynamicStorageBuffersPerPipelineLayout: l.MaxDynamicStorageBuffersPerPipelineLayout,
		MaxSampledTexturesPerShaderStage:          l.MaxSampledTexturesPerShaderStage,
		MaxSamplersPerShaderStage:                 l.MaxSamplersPerShaderStage,
		MaxStorageBuffersPerShaderStage:           l.MaxStorageBuffersPerShaderStage,
		MaxStorageTexturesPerShaderStage:          l.MaxStorageTexturesPerShaderStage,
		MaxUniformBuffersPerShaderStage:           l.MaxUniformBuffersPerShaderStage,
		MaxUniformBufferBindingSize:               l.MaxUniformBufferBindingSize,
		MaxStorageBufferBindingSize:               l.MaxStorageBufferBindingSize,
		MinUniformBufferOffsetAlignment:           l.MinUniformBufferOffsetAlignment,
		MinStorageBufferOffsetAlignment:           l.MinStorageBufferOffsetAlignment,
		MaxVertexBuffers:                          l.MaxVertexBuffers,
		MaxBufferSize:                             l.MaxBufferSize,
		MaxVertexAttributes:                       l.MaxVertexAttributes,
		MaxVertexBufferArrayStride:                l.MaxVertexBufferArrayStride,
		MaxInterStageShaderVariables:              l.MaxInterStageShaderVariables,
		MaxColorAttachments:                       l.MaxColorAttachments,
		MaxColorAttachmentBytesPerSample:          l.MaxColorAttachmentBytesPerSample,
		MaxComputeWorkgroupStorageSize:            l.MaxComputeWorkgroupStorageSize,
		MaxComputeInvocationsPerWorkgroup:         l.MaxComputeInvocationsPerWorkgroup,
		MaxComputeWorkgroupSizeX:                  l.MaxComputeWorkgroupSizeX,
		MaxComputeWorkgroupSizeY:                  l.MaxComputeWorkgroupSizeY,
		MaxComputeWorkgroupSizeZ:                  l.MaxComputeWorkgroupSizeZ,
		MaxComputeWorkgroupsPerDimension:          l.MaxComputeWorkgroupsPerDimension,
	}
}

// deviceDescriptorWire is the FFI-compatible C-layout struct for wgpuAdapterRequestDevice.
// v29: Added Label, RequiredFeatureCount, RequiredFeatures, RequiredLimits,
// DefaultQueue, DeviceLostCallbackInfo, UncapturedErrorCallbackInfo fields.
type deviceDescriptorWire struct {
	NextInChain                 uintptr // *ChainedStruct
	Label                       StringView
	RequiredFeatureCount        uintptr // size_t
	RequiredFeatures            uintptr // *FeatureName (const)
	RequiredLimits              uintptr // *Limits (const, nullable)
	DefaultQueue                QueueDescriptor
	DeviceLostCallbackInfo      DeviceLostCallbackInfo
	UncapturedErrorCallbackInfo UncapturedErrorCallbackInfo
}

// QueueDescriptor configures queue creation.
type QueueDescriptor struct {
	NextInChain uintptr // *ChainedStruct
	Label       StringView
}

// CreateDepthTexture creates a depth texture with the specified dimensions and format.
// This is a convenience function for creating depth buffers for render passes.
// Returns nil on error (use CreateTexture directly for full error handling).
func (d *Device) CreateDepthTexture(width, height uint32, format gputypes.TextureFormat) *Texture {
	desc := TextureDescriptor{
		Usage:         gputypes.TextureUsageRenderAttachment,
		Dimension:     gputypes.TextureDimension2D,
		Size:          gputypes.Extent3D{Width: width, Height: height, DepthOrArrayLayers: 1},
		Format:        format,
		MipLevelCount: 1,
		SampleCount:   1,
	}

	t, _ := d.CreateTexture(&desc)
	return t
}

// Limits returns the resource limits of this device.
//
// Limits are cached at device creation time and returned by value.
// No FFI call is made. Returns zero-value Limits if the device is nil.
// This matches the gogpu/wgpu API signature for cross-project compatibility.
func (d *Device) Limits() Limits {
	if d == nil || d.handle == 0 {
		return Limits{}
	}
	return d.limits
}

// Features retrieves all features enabled on this device.
// Returns a slice of FeatureName values.
func (d *Device) Features() []FeatureName {
	mustInit()
	if d == nil || d.handle == 0 {
		return nil
	}

	// Call wgpuDeviceGetFeatures to populate SupportedFeatures struct
	var supported SupportedFeatures
	wgpunative.DeviceGetFeatures(
		d.handle,
		uintptr(unsafe.Pointer(&supported)),
	)

	if supported.FeatureCount == 0 || supported.Features == 0 {
		return nil
	}

	// Convert C array to Go slice
	featuresPtr := (*FeatureName)(ptrFromUintptr(supported.Features))
	features := unsafe.Slice(featuresPtr, supported.FeatureCount)

	// Copy to new slice (don't keep pointer to C memory)
	result := make([]FeatureName, supported.FeatureCount)
	copy(result, features)

	// Free C-allocated memory (pass pointer to struct, not individual fields)
	wgpunative.SupportedFeaturesFreeMembers(uintptr(unsafe.Pointer(&supported)))

	return result
}

// HasFeature checks if the device has a specific feature enabled.
func (d *Device) HasFeature(feature FeatureName) bool {
	mustInit()
	if d == nil || d.handle == 0 {
		return false
	}

	result := wgpunative.DeviceHasFeature(
		d.handle,
		uintptr(feature),
	)

	return Bool(result) == True
}
