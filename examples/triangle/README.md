# WGPU native triangle example

This is the Go port of `third_party/wgpu-native/examples/triangle`. It creates
a native Win32 window, selects an adapter, creates a render pipeline from the
WGSL triangle shader, and presents a red triangle continuously.

The example is Windows amd64-specific because it uses a native HWND. Build it
from the repository root after the static archive has been built:

```powershell
$env:CGO_ENABLED = "1"
$env:CC = "C:\local\gnu\msys64\ucrt64\bin\clang.exe"
go run .
```

The `.syso` package is imported so the native archive participates in the
executable link. No runtime DLL path or GoFFI loader is used.
