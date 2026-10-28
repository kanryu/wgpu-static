#define WIN32_LEAN_AND_MEAN
#include <windows.h>

#include "webgpu-headers/webgpu.h"
#include "wgpu.h"

#include <stdio.h>

static const wchar_t kClassName[] = L"WGPUStaticTriangleC";

static void log_callback(WGPULogLevel level, WGPUStringView message,
                         void *userdata) {
  (void)userdata;
  printf("wgpu-log level=%u: %.*s\n", (unsigned)level,
         (int)message.length, message.data ? message.data : "");
}

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
  WNDCLASSEXW window_class = {0};
  window_class.cbSize = sizeof(window_class);
  window_class.hInstance = instance;
  window_class.lpfnWndProc = window_proc;
  window_class.lpszClassName = kClassName;
  window_class.style = CS_HREDRAW | CS_VREDRAW;

  if (RegisterClassExW(&window_class) == 0 &&
      GetLastError() != ERROR_CLASS_ALREADY_EXISTS) {
    fprintf(stderr, "RegisterClassExW failed: %lu\n", GetLastError());
    return NULL;
  }

  HWND hwnd = CreateWindowExW(0, kClassName, L"triangle [C + wgpu-native]",
                              WS_OVERLAPPEDWINDOW | WS_VISIBLE, 100, 100, 640,
                              480, NULL, NULL, instance, NULL);
  if (hwnd == NULL) {
    fprintf(stderr, "CreateWindowExW failed: %lu\n", GetLastError());
  }
  return hwnd;
}

int main(void) {
  HINSTANCE instance_handle = GetModuleHandleW(NULL);
  HWND hwnd = create_window(instance_handle);
  if (hwnd == NULL) {
    return 1;
  }

  printf("setting wgpu log level: trace\n");
  wgpuSetLogCallback(log_callback, NULL);
  wgpuSetLogLevel(WGPULogLevel_Trace);
  printf("creating instance\n");
  WGPUInstance instance = wgpuCreateInstance(NULL);
  if (instance == NULL) {
    fprintf(stderr, "wgpuCreateInstance failed\n");
    DestroyWindow(hwnd);
    return 2;
  }

  WGPUSurfaceSourceWindowsHWND source = {
      .chain =
          {
              .next = NULL,
              .sType = WGPUSType_SurfaceSourceWindowsHWND,
          },
      .hinstance = instance_handle,
      .hwnd = hwnd,
  };
  WGPUSurfaceDescriptor descriptor = {
      .nextInChain = (const WGPUChainedStruct *)&source,
      .label = WGPU_STRING_VIEW_INIT,
  };

  printf("creating surface\n");
  WGPUSurface surface = wgpuInstanceCreateSurface(instance, &descriptor);
  if (surface == NULL) {
    fprintf(stderr, "wgpuInstanceCreateSurface failed\n");
    wgpuInstanceRelease(instance);
    DestroyWindow(hwnd);
    return 3;
  }

  printf("surface created successfully\n");
  wgpuSurfaceRelease(surface);
  wgpuInstanceRelease(instance);
  DestroyWindow(hwnd);
  return 0;
}
