# WGPU native static-link example

This example imports the generated Hike bindings and blank-imports the package
that carries `wgpunative_native_windows_amd64.syso`. The executable therefore
resolves `wgpuGetVersion` from the statically linked `wgpu-native` archive.

Run on Windows amd64 from a UCRT64/MSYS2 environment:

```powershell
$env:CGO_ENABLED = "1"
$env:CC = "C:\local\gnu\msys64\ucrt64\bin\clang.exe"
$env:CXX = "C:\local\gnu\msys64\ucrt64\bin\clang++.exe"
go run .
```

The package name is `0.0.0`, so the native version value is `0`:

```text
wgpu-native version: 0
```
