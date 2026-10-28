# Pure Go triangle: wgpu-native API disassembly notes

This note records the API entry points used by the Pure Go diagnostic triangle
program. The debug DLL was disassembled with `llvm-objdump`, and the program
was launched under LLDB with one-shot breakpoints on the exported WGPU entry
points. The addresses below are RVAs as printed by the debug DLL's symbol
table; they can change when the DLL is rebuilt.

The Windows x64 entry points receive their first four integer/pointer
arguments in `RCX`, `RDX`, `R8`, and `R9`. The observations below are about the
native Rust/C ABI entry points, not about the Go `syscall.Proc.Call` wrapper.

## Observed call order

The non-release calls were entered in this order during the debug run:

1. `wgpuSetLogLevel`
2. `wgpuGetVersion`
3. `wgpuCreateInstance`
4. `wgpuInstanceCreateSurface`
5. `wgpuInstanceRequestAdapter`
6. `wgpuAdapterRequestDevice`
7. `wgpuDeviceGetQueue`
8. `wgpuSurfaceGetCapabilities`
9. `wgpuSurfaceConfigure`
10. `wgpuSurfaceCapabilitiesFreeMembers`
11. `wgpuDeviceCreateShaderModule`
12. `wgpuDeviceCreateRenderPipeline`
13. `wgpuSurfaceGetCurrentTexture`
14. `wgpuTextureCreateView`
15. `wgpuDeviceCreateCommandEncoder`
16. `wgpuCommandEncoderBeginRenderPass`
17. `wgpuRenderPassEncoderSetPipeline`
18. `wgpuRenderPassEncoderDraw`
19. `wgpuRenderPassEncoderEnd`
20. `wgpuCommandEncoderFinish`
21. `wgpuQueueSubmit`
22. `wgpuSurfacePresent`

`wgpuInstanceProcessEvents` is entered repeatedly while waiting for the
adapter/device callbacks and once per render-loop iteration.

The release calls run during shutdown or at the end of a frame:
`wgpuTextureViewRelease`, `wgpuTextureRelease`, `wgpuCommandBufferRelease`,
`wgpuRenderPassEncoderRelease`, `wgpuCommandEncoderRelease`,
`wgpuRenderPipelineRelease`, `wgpuShaderModuleRelease`, `wgpuQueueRelease`,
`wgpuDeviceRelease`, `wgpuAdapterRelease`, `wgpuSurfaceRelease`, and
`wgpuInstanceRelease`.

## Entry-point notes

| API | RVA | Entry stack operation | What the entry point does before the backend call |
|---|---:|---|---|
| `wgpuSetLogLevel` | `0x47080` | `sub rsp, 0x38` | Converts the public log-level enum and calls the native logging max-level setter. |
| `wgpuGetVersion` | `0x46bd0` | `sub rsp, 0x258` | Builds the version result and returns it as a 32-bit value. |
| `wgpuCreateInstance` | `0x19180` | `sub rsp, 0x768` | Checks the optional descriptor, validates chained data, maps the instance descriptor, and constructs the global instance. |
| `wgpuInstanceCreateSurface` | `0x205d0` | `push rsi; push rdi; sub rsp, 0x718` | Validates the instance, walks the surface descriptor chain, maps the Windows HWND source, and creates the surface. |
| `wgpuInstanceProcessEvents` | `0x219f0` | `sub rsp, 0xb8` | Validates the instance and services queued futures/callback events. |
| `wgpuInstanceRequestAdapter` | `0x21b80` | `sub rsp, 0x278` | Validates the instance, maps adapter options and callback information, submits the request, and returns a future handle. |
| `wgpuAdapterRequestDevice` | `0x127d0` | `sub rsp, 0xea8` | Validates the adapter, maps device and callback descriptors, starts device creation, and returns a future handle. |
| `wgpuDeviceGetQueue` | `0x1f7f0` | `sub rsp, 0x98` | Validates the device and obtains its queue handle. |
| `wgpuSurfaceGetCapabilities` | `0x29f50` | `sub rsp, 0x458` | Validates surface/adapter/output pointers, queries backend capabilities, and writes the output structure. |
| `wgpuSurfaceCapabilitiesFreeMembers` | `0x29740` | `sub rsp, 0x98` | Reads the capability-array pointers and releases the native capability members. |
| `wgpuSurfaceConfigure` | `0x29870` | `sub rsp, 0x2c8` | Validates the surface, maps the configuration, and configures the swapchain surface. |
| `wgpuDeviceCreateShaderModule` | `0x1e420` | `sub rsp, 0x78` | Applies default shader runtime checks, maps the WGSL chained source, and creates the shader module. |
| `wgpuDeviceCreateRenderPipeline` | `0x1c930` | `sub rsp, 0x8e8` | Validates the device, maps vertex/primitive/multisample/fragment state, then creates the pipeline. |
| `wgpuSurfaceGetCurrentTexture` | `0x2aa70` | `sub rsp, 0x258` | Validates the surface and output pointer, acquires the next surface texture, and writes texture/status. |
| `wgpuTextureCreateView` | `0x2b5e0` | `sub rsp, 0x2f8` | Validates the texture, maps the optional view descriptor, and creates a texture view. |
| `wgpuDeviceCreateCommandEncoder` | `0x1a890` | `sub rsp, 0x1c8` | Validates the device, maps the optional descriptor, and creates a command encoder. |
| `wgpuCommandEncoderBeginRenderPass` | `0x14a10` | `mov eax, 0x1148; call __chkstk_ms; sub rsp, rax` | Validates the encoder, maps color attachments and pass state, and begins the render pass. |
| `wgpuRenderPassEncoderSetPipeline` | `0x28170` | `sub rsp, 0xe8` | Validates the render-pass encoder and pipeline, then records the pipeline binding. |
| `wgpuRenderPassEncoderDraw` | `0x25460` | `sub rsp, 0xd8` | Copies the four 32-bit draw arguments from the ABI registers/stack and records the draw command. |
| `wgpuRenderPassEncoderEnd` | `0x25d80` | `sub rsp, 0xb8` | Validates the pass encoder and closes the pass. |
| `wgpuCommandEncoderFinish` | `0x16360` | `sub rsp, 0x3e8` | Validates the encoder, maps the optional descriptor, finishes the command buffer, and returns it. |
| `wgpuQueueSubmit` | `0x22960` | `sub rsp, 0x2c8` | Validates the queue, copies the command-buffer slice, and submits it to the device queue. |
| `wgpuSurfacePresent` | `0x2b0e0` | `sub rsp, 0x148` | Validates the surface and presents the acquired frame. |

## Release-entry pattern

All release entry points use the same initial pattern: reserve `0x48` bytes,
save the handle from `RCX`, call the corresponding Rust null-check helper, and
then enter the reference-count/deallocation path. The observed entry points
are:

| API | RVA |
|---|---:|
| `wgpuTextureViewRelease` | `0x2c170` |
| `wgpuTextureRelease` | `0x2c070` |
| `wgpuCommandBufferRelease` | `0x142a0` |
| `wgpuRenderPassEncoderRelease` | `0x27680` |
| `wgpuCommandEncoderRelease` | `0x16da0` |
| `wgpuRenderPipelineRelease` | `0x29350` |
| `wgpuShaderModuleRelease` | `0x29550` |
| `wgpuQueueRelease` | `0x228e0` |
| `wgpuDeviceRelease` | `0x20200` |
| `wgpuAdapterRelease` | `0x12750` |
| `wgpuSurfaceRelease` | `0x2b3b0` |
| `wgpuInstanceRelease` | `0x21b00` |

## ABI-relevant findings

- The native entry points preserve the Windows x64 register argument order and
  reserve their own stack space internally. A caller does not need to predict
  the large internal Rust stack frames such as `0x768`, `0x8e8`, or `0x1148`.
- `wgpuRenderPassEncoderDraw` has four 32-bit integer arguments after the
  encoder handle. The fourth argument is read from the stack shadow/argument
  area after the register arguments have been saved.
- `wgpuInstanceRequestAdapter` and `wgpuAdapterRequestDevice` receive callback
  information as a by-value structure in `R8`/the following ABI area, not as a
  pointer to a Go-managed callback structure. This is a critical comparison
  point for the Hike/SXSO trampoline.
- `wgpuSurfaceGetCurrentTexture` and `wgpuSurfaceGetCapabilities` write output
  structures through pointers supplied by the caller; these are not returned
  in `RAX`.
- The Pure Go path reaches all listed entries and renders frames. Therefore,
  a static-link failure in the same sequence is evidence against the WGPU
  backend itself and in favor of a wrapper/trampoline ABI mismatch.
