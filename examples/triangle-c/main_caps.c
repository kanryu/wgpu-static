#define WIN32_LEAN_AND_MEAN
#include <windows.h>

#include "webgpu-headers/webgpu.h"
#include "wgpu.h"

#include <stdio.h>

typedef struct {
  WGPUAdapter adapter;
  WGPUDevice device;
  int adapter_done;
  int device_done;
} state_t;

static LRESULT CALLBACK window_proc(HWND hwnd, UINT message, WPARAM wparam,
                                    LPARAM lparam) {
  (void)wparam;
  (void)lparam;
  if (message == WM_DESTROY) {
    PostQuitMessage(0);
    return 0;
  }
  return DefWindowProcW(hwnd, message, wparam, lparam);
}

static HWND create_window(HINSTANCE instance) {
  static const wchar_t class_name[] = L"WGPUStaticCapsC";
  WNDCLASSEXW wc = {0};
  wc.cbSize = sizeof(wc);
  wc.hInstance = instance;
  wc.lpfnWndProc = window_proc;
  wc.lpszClassName = class_name;
  if (RegisterClassExW(&wc) == 0 && GetLastError() != ERROR_CLASS_ALREADY_EXISTS)
    return NULL;
  return CreateWindowExW(0, class_name, L"wgpu C capabilities probe",
                         WS_OVERLAPPEDWINDOW | WS_VISIBLE, 100, 100, 640, 480,
                         NULL, NULL, instance, NULL);
}

static void adapter_callback(WGPURequestAdapterStatus status,
                             WGPUAdapter adapter, WGPUStringView message,
                             void *userdata1, void *userdata2) {
  state_t *state = userdata1;
  (void)userdata2;
  state->adapter = adapter;
  state->adapter_done = 1;
  printf("C adapter callback: status=%u message=%.*s\n", (unsigned)status,
         (int)message.length, message.data ? message.data : "");
}

static void device_callback(WGPURequestDeviceStatus status, WGPUDevice device,
                            WGPUStringView message, void *userdata1,
                            void *userdata2) {
  state_t *state = userdata1;
  (void)userdata2;
  state->device = device;
  state->device_done = 1;
  printf("C device callback: status=%u message=%.*s\n", (unsigned)status,
         (int)message.length, message.data ? message.data : "");
}

int main(void) {
  HINSTANCE hinstance = GetModuleHandleW(NULL);
  HWND hwnd = create_window(hinstance);
  if (!hwnd) return 1;

  printf("C creating instance\n");
  WGPUInstance instance = wgpuCreateInstance(NULL);
  if (!instance) return 2;

  WGPUSurfaceSourceWindowsHWND source = {
      .chain = {.next = NULL, .sType = WGPUSType_SurfaceSourceWindowsHWND},
      .hinstance = hinstance,
      .hwnd = hwnd,
  };
  WGPUSurfaceDescriptor surface_desc = {
      .nextInChain = (WGPUChainedStruct *)&source,
      .label = WGPU_STRING_VIEW_INIT,
  };
  printf("C creating surface\n");
  WGPUSurface surface = wgpuInstanceCreateSurface(instance, &surface_desc);
  if (!surface) return 3;

  state_t state = {0};
  printf("C requesting adapter\n");
  wgpuInstanceRequestAdapter(
      instance,
      &(const WGPURequestAdapterOptions){.compatibleSurface = surface},
      (WGPURequestAdapterCallbackInfo){
          .mode = WGPUCallbackMode_AllowProcessEvents,
          .callback = adapter_callback,
          .userdata1 = &state,
      });
  while (!state.adapter_done) wgpuInstanceProcessEvents(instance);

  printf("C requesting device\n");
  wgpuAdapterRequestDevice(
      state.adapter, NULL,
      (WGPURequestDeviceCallbackInfo){
          .mode = WGPUCallbackMode_AllowProcessEvents,
          .callback = device_callback,
          .userdata1 = &state,
      });
  while (!state.device_done) wgpuInstanceProcessEvents(instance);

  WGPUSurfaceCapabilities capabilities = WGPU_SURFACE_CAPABILITIES_INIT;
  printf("C getting surface capabilities\n");
  WGPUStatus status =
      wgpuSurfaceGetCapabilities(surface, state.adapter, &capabilities);
  printf("C capabilities returned: status=%u formats=%zu present=%zu alpha=%zu\n",
         (unsigned)status, capabilities.formatCount,
         capabilities.presentModeCount, capabilities.alphaModeCount);

  WGPUSurfaceConfiguration config = WGPU_SURFACE_CONFIGURATION_INIT;
  config.device = state.device;
  config.format = capabilities.formats[0];
  config.width = 640;
  config.height = 480;
  config.alphaMode = capabilities.alphaModes[0];
  config.presentMode = capabilities.presentModes[0];
  printf("C configuring surface\n");
  wgpuSurfaceConfigure(surface, &config);
  printf("C surface configured\n");

  printf("C getting queue\n");
  WGPUQueue queue = wgpuDeviceGetQueue(state.device);
  const char shader_code[] =
      "@vertex\n"
      "fn vs_main(@builtin(vertex_index) index: u32) -> @builtin(position) vec4<f32> {\n"
      "    let x = f32(i32(index) - 1);\n"
      "    let y = f32(i32(index & 1u) * 2 - 1);\n"
      "    return vec4<f32>(x, y, 0.0, 1.0);\n"
      "}\n"
      "@fragment\n"
      "fn fs_main() -> @location(0) vec4<f32> {\n"
      "    return vec4<f32>(1.0, 0.0, 0.0, 1.0);\n"
      "}";
  WGPUShaderSourceWGSL shader_source = {
      .chain = {.next = NULL, .sType = WGPUSType_ShaderSourceWGSL},
      .code = {shader_code, sizeof(shader_code) - 1},
  };
  WGPUShaderModule shader = wgpuDeviceCreateShaderModule(
      state.device,
      &(WGPUShaderModuleDescriptor){
          .nextInChain = (WGPUChainedStruct *)&shader_source,
          .label = WGPU_STRING_VIEW_INIT,
      });
  printf("C shader module=%p\n", (void *)shader);
  WGPUPipelineLayout layout = wgpuDeviceCreatePipelineLayout(
      state.device,
      &(WGPUPipelineLayoutDescriptor){.label = WGPU_STRING_VIEW_INIT});
  printf("C pipeline layout=%p queue=%p\n", (void *)layout, (void *)queue);

  WGPUVertexState vertex = {
      .module = shader,
      .entryPoint = {"vs_main", 7},
  };
  WGPUPrimitiveState primitive = WGPU_PRIMITIVE_STATE_INIT;
  primitive.topology = WGPUPrimitiveTopology_TriangleList;
  primitive.frontFace = WGPUFrontFace_CCW;
  primitive.cullMode = WGPUCullMode_None;
  WGPUMultisampleState multisample = WGPU_MULTISAMPLE_STATE_INIT;
  WGPUColorTargetState target = {
      .format = capabilities.formats[0],
      .writeMask = WGPUColorWriteMask_All,
  };
  WGPUFragmentState fragment = {
      .module = shader,
      .entryPoint = {"fs_main", 7},
      .targetCount = 1,
      .targets = &target,
  };
  WGPURenderPipeline pipeline = wgpuDeviceCreateRenderPipeline(
      state.device,
      &(WGPURenderPipelineDescriptor){
          .layout = layout,
          .vertex = vertex,
          .fragment = &fragment,
          .primitive = primitive,
          .multisample = multisample,
      });
  printf("C render pipeline=%p\n", (void *)pipeline);

  WGPUSurfaceTexture surface_texture = WGPU_SURFACE_TEXTURE_INIT;
  printf("C calling SurfaceGetCurrentTexture\n");
  fflush(stdout);
  wgpuSurfaceGetCurrentTexture(surface, &surface_texture);
  printf("C returned SurfaceGetCurrentTexture status=%u\n",
         (unsigned)surface_texture.status);
  fflush(stdout);
  printf("C calling TextureCreateView\n");
  fflush(stdout);
  WGPUTextureView view = wgpuTextureCreateView(surface_texture.texture, NULL);
  printf("C returned TextureCreateView=%p\n", (void *)view);
  fflush(stdout);
  printf("C calling DeviceCreateCommandEncoder\n");
  fflush(stdout);
  WGPUCommandEncoder encoder =
      wgpuDeviceCreateCommandEncoder(state.device, NULL);
  printf("C returned DeviceCreateCommandEncoder=%p\n", (void *)encoder);
  fflush(stdout);
  WGPURenderPassColorAttachment color = {
      .view = view,
      .depthSlice = WGPU_DEPTH_SLICE_UNDEFINED,
      .resolveTarget = NULL,
      .loadOp = WGPULoadOp_Clear,
      .storeOp = WGPUStoreOp_Store,
      .clearValue = {0.0, 0.0, 0.0, 1.0},
  };
  WGPURenderPassEncoder pass = wgpuCommandEncoderBeginRenderPass(
      encoder,
      &(WGPURenderPassDescriptor){.colorAttachmentCount = 1,
                                  .colorAttachments = &color});
  printf("C returned CommandEncoderBeginRenderPass=%p\n", (void *)pass);
  fflush(stdout);
  wgpuRenderPassEncoderSetPipeline(pass, pipeline);
  printf("C returned RenderPassEncoderSetPipeline\n");
  fflush(stdout);
  wgpuRenderPassEncoderDraw(pass, 3, 1, 0, 0);
  printf("C returned RenderPassEncoderDraw\n");
  fflush(stdout);
  wgpuRenderPassEncoderEnd(pass);
  printf("C returned RenderPassEncoderEnd\n");
  fflush(stdout);
  wgpuRenderPassEncoderRelease(pass);
  WGPUCommandBuffer command_buffer = wgpuCommandEncoderFinish(encoder, NULL);
  printf("C returned CommandEncoderFinish=%p\n", (void *)command_buffer);
  fflush(stdout);
  wgpuQueueSubmit(queue, 1, &command_buffer);
  printf("C returned QueueSubmit\n");
  fflush(stdout);
  wgpuSurfacePresent(surface);
  printf("C returned SurfacePresent\n");
  fflush(stdout);
  wgpuCommandBufferRelease(command_buffer);
  wgpuCommandEncoderRelease(encoder);
  wgpuTextureViewRelease(view);
  wgpuTextureRelease(surface_texture.texture);
  wgpuRenderPipelineRelease(pipeline);

  wgpuSurfaceCapabilitiesFreeMembers(capabilities);
  wgpuPipelineLayoutRelease(layout);
  wgpuShaderModuleRelease(shader);
  wgpuQueueRelease(queue);
  wgpuDeviceRelease(state.device);
  wgpuAdapterRelease(state.adapter);
  wgpuSurfaceRelease(surface);
  wgpuInstanceRelease(instance);
  DestroyWindow(hwnd);
  return 0;
}
