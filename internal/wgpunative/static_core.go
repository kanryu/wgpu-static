//go:build ignore

// Package wgpunative exposes the Hike/SXSO-linked wgpu-native entry points.
package wgpunative

func CreateInstance(descriptor *byte) *byte { return _hike_CreateInstance(descriptor) }
func InstanceRelease(instance *byte) { _hike_InstanceRelease(instance) }
func InstanceProcessEvents(instance *byte) { _hike_InstanceProcessEvents(instance) }
func InstanceCreateSurface(instance *byte, descriptor *byte) *byte { return _hike_InstanceCreateSurface(instance, descriptor) }
func InstanceRequestAdapter(instance *byte, options *byte, callbackInfo *byte) uint64 { return _hike_InstanceRequestAdapter(instance, options, callbackInfo) }
func AdapterRelease(adapter *byte) { _hike_AdapterRelease(adapter) }
func AdapterRequestDevice(adapter *byte, descriptor *byte, callbackInfo *byte) uint64 { return _hike_AdapterRequestDevice(adapter, descriptor, callbackInfo) }
func DeviceRelease(device *byte) { _hike_DeviceRelease(device) }
func DeviceGetQueue(device *byte) *byte { return _hike_DeviceGetQueue(device) }
func DeviceCreateShaderModule(device *byte, descriptor *byte) *byte { return _hike_DeviceCreateShaderModule(device, descriptor) }
func ShaderModuleRelease(shaderModule *byte) { _hike_ShaderModuleRelease(shaderModule) }
func DeviceCreatePipelineLayout(device *byte, descriptor *byte) *byte { return _hike_DeviceCreatePipelineLayout(device, descriptor) }
func PipelineLayoutRelease(layout *byte) { _hike_PipelineLayoutRelease(layout) }
func SurfaceGetCapabilities(surface, adapter, capabilities *byte) uint32 { return _hike_SurfaceGetCapabilities(surface, adapter, capabilities) }
func SurfaceCapabilitiesFreeMembers(capabilities *byte) { _hike_SurfaceCapabilitiesFreeMembers(capabilities) }
func SurfaceConfigure(surface, config *byte) { _hike_SurfaceConfigure(surface, config) }
func SurfaceGetCurrentTexture(surface, texture *byte) { _hike_SurfaceGetCurrentTexture(surface, texture) }
func SurfacePresent(surface *byte) { _hike_SurfacePresent(surface) }
func SurfaceRelease(surface *byte) { _hike_SurfaceRelease(surface) }
func TextureCreateView(texture, descriptor *byte) *byte { return _hike_TextureCreateView(texture, descriptor) }
func TextureViewRelease(view *byte) { _hike_TextureViewRelease(view) }
func TextureRelease(texture *byte) { _hike_TextureRelease(texture) }
func DeviceCreateRenderPipeline(device, descriptor *byte) *byte { return _hike_DeviceCreateRenderPipeline(device, descriptor) }
func RenderPipelineRelease(pipeline *byte) { _hike_RenderPipelineRelease(pipeline) }
func DeviceCreateCommandEncoder(device, descriptor *byte) *byte { return _hike_DeviceCreateCommandEncoder(device, descriptor) }
func CommandEncoderBeginRenderPass(encoder, descriptor *byte) *byte { return _hike_CommandEncoderBeginRenderPass(encoder, descriptor) }
func RenderPassEncoderSetPipeline(pass, pipeline *byte) { _hike_RenderPassEncoderSetPipeline(pass, pipeline) }
func RenderPassEncoderDraw(pass *byte, vertexCount, instanceCount, firstVertex, firstInstance uint32) { _hike_RenderPassEncoderDraw(pass, vertexCount, instanceCount, firstVertex, firstInstance) }
func RenderPassEncoderEnd(pass *byte) { _hike_RenderPassEncoderEnd(pass) }
func RenderPassEncoderRelease(pass *byte) { _hike_RenderPassEncoderRelease(pass) }
func CommandEncoderFinish(encoder, descriptor *byte) *byte { return _hike_CommandEncoderFinish(encoder, descriptor) }
func CommandEncoderRelease(encoder *byte) { _hike_CommandEncoderRelease(encoder) }
func CommandBufferRelease(buffer *byte) { _hike_CommandBufferRelease(buffer) }
func QueueSubmit(queue *byte, commandCount uint64, commands *byte) { _hike_QueueSubmit(queue, commandCount, commands) }
func QueueRelease(queue *byte) { _hike_QueueRelease(queue) }
func GetVersion() uint32 { return _hike_GetVersion() }

func _hike_CreateInstance(*byte) *byte
func _hike_InstanceRelease(*byte)
func _hike_InstanceProcessEvents(*byte)
func _hike_InstanceCreateSurface(*byte, *byte) *byte
func _hike_InstanceRequestAdapter(*byte, *byte, *byte) uint64
func _hike_AdapterRelease(*byte)
func _hike_AdapterRequestDevice(*byte, *byte, *byte) uint64
func _hike_DeviceRelease(*byte)
func _hike_DeviceGetQueue(*byte) *byte
func _hike_DeviceCreateShaderModule(*byte, *byte) *byte
func _hike_ShaderModuleRelease(*byte)
func _hike_DeviceCreatePipelineLayout(*byte, *byte) *byte
func _hike_PipelineLayoutRelease(*byte)
func _hike_SurfaceGetCapabilities(*byte, *byte, *byte) uint32
func _hike_SurfaceCapabilitiesFreeMembers(*byte)
func _hike_SurfaceConfigure(*byte, *byte)
func _hike_SurfaceGetCurrentTexture(*byte, *byte)
func _hike_SurfacePresent(*byte)
func _hike_SurfaceRelease(*byte)
func _hike_TextureCreateView(*byte, *byte) *byte
func _hike_TextureViewRelease(*byte)
func _hike_TextureRelease(*byte)
func _hike_DeviceCreateRenderPipeline(*byte, *byte) *byte
func _hike_RenderPipelineRelease(*byte)
func _hike_DeviceCreateCommandEncoder(*byte, *byte) *byte
func _hike_CommandEncoderBeginRenderPass(*byte, *byte) *byte
func _hike_RenderPassEncoderSetPipeline(*byte, *byte)
func _hike_RenderPassEncoderDraw(*byte, uint32, uint32, uint32, uint32)
func _hike_RenderPassEncoderEnd(*byte)
func _hike_RenderPassEncoderRelease(*byte)
func _hike_CommandEncoderFinish(*byte, *byte) *byte
func _hike_CommandEncoderRelease(*byte)
func _hike_CommandBufferRelease(*byte)
func _hike_QueueSubmit(*byte, uint64, *byte)
func _hike_QueueRelease(*byte)
func _hike_GetVersion() uint32
