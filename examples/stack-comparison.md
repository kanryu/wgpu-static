# C and minimal-Go stack comparison

The comparison uses the same debug `libwgpu_native.a` and stops immediately
after the native API returns where possible.

## C version

Breakpoint: `examples/triangle-c/main.c:76`, immediately after
`wgpuInstanceCreateSurface(instance, &descriptor)`.

```text
* thread #1, stop reason = breakpoint
  * frame #0  triangle-c-debug.exe`main at main.c:76:15
    frame #1  triangle-c-debug.exe`main startup frame
    frame #2  kernel32.dll`BaseThreadInitThunk
    frame #3  ntdll.dll`RtlUserThreadStart
```

The call returned normally. The C program printed:

```text
creating instance
creating surface
surface created successfully
exit=0
```

## Minimal Go version

The minimal program calls `wgpunative.GetVersion`,
`wgpunative.CreateInstance`, and `wgpunative.InstanceCreateSurface` directly.
It does not use the `wgpu.Instance` wrapper, callback initialization, resource
tracking, or error conversion.

With LLDB attached, the `wgpuCreateInstance` call returns and the post-call
stack is:

```text
* thread #1, stop reason = breakpoint
  * frame #0  surface-minimal-debug.exe`main.main
    at main_windows_amd64.go:101
    frame #1  runtime.main at proc.go:302
    frame #2  runtime.goexit.abi0 at asm_amd64.s:1264
```

Without LLDB, the same call stops before returning:

```text
Exception 0x40010006
runtime.winthrow
runtime.lastcontinuehandler
runtime.sigtrampgo
runtime.sigtramp
```

`0x40010006` is the Windows `DBG_PRINTEXCEPTION_C` debug-output exception.
The native C program lets this notification pass through normally; the Go
runtime's Windows exception handler treats it as an unhandled exception and
enters `runtime.winthrow`.

Starting LLDB with:

```text
log enable windows process
settings set target.process.stop-on-sharedlibrary-events false
run
```

suppresses the debugger stop and allows both `wgpuCreateInstance` and
`wgpuInstanceCreateSurface` to return. It does not make the program correct:
the test subsequently exits with a separate `0xc0000005` access violation.
Calling `wgpuSetLogLevel(WGPULogLevel_Off)` before instance creation also does
not remove the `0x40010006` notification, so it is not caused by wgpu's own
log-level callback.

With both implementations configured for `WGPULogLevel_Trace`, the C callback
receives Vulkan loader and backend messages and the C program exits normally.
A Go callback installed through `syscall.NewCallback` fails before the first
WGPU log message with:

```text
fatal error: exitsyscall: syscall frame is no longer valid
runtime.cgocallback
runtime.cgocall
```

This is a separate boundary issue: the SXSO trampoline calls the native
function directly and does not establish the Go runtime's cgo syscall frame
required by `syscall.NewCallback`. It is not evidence of a bad surface
descriptor.

## Conclusion

The post-call C stack is a normal return to C. The minimal Go stack is also a
normal return when a debugger consumes the debug-output exception, but without
a debugger the notification is converted into a Go runtime failure. The
failure occurs before the surface descriptor is relevant in the minimal test,
so the high-level Go wrapper and the surface parameter layout are not the
cause.
