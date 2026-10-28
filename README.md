# WGPUStatic

WGPUStatic is a Go extension project for binding the Rust implementation of
[`wgpu-native`](https://github.com/gfx-rs/wgpu-native) through a statically
linked native library.

The project keeps the Go-facing WGPU interface close to the existing FFI-based
WebGPU binding. The `wgpu` package source is based on the source used by the Go
WebGPU FFI project, while native calls are provided by the Hike/SXSO-generated
`internal/wgpunative` package. The public package has no GoFFI dependency.

The main difference from the original FFI project is the native link model:

| Project | Go API and FFI | Native WGPU library |
|---|---|---|
| Original FFI project | Pure-Go FFI with the existing WGPU binding | Loaded dynamically at runtime |
| WGPUStatic | The same FFI-oriented WGPU binding sources | Rust `wgpu-native` linked into the executable as a static archive |

## Architecture

```text
Go application
    │
    ├── wgpu package
    │       └── Go WebGPU binding adapted for static native calls
    │
    ├── internal/wgpunative
    │       └── Hike declarations and native symbol stubs
    │
    └── internal/wgpunative_native
            └── wgpunative_native_windows_amd64.syso
                    │
                    └── libwgpu_native.a
                            │
                            └── Rust wgpu-native v29.0.1.1
```

The native repository is registered as the `third_party/wgpu-native` Git
submodule and is fixed to tag `v29.0.1.1`. Its Cargo configuration produces a
static library in addition to the shared-library target. The resulting archive
is copied into `internal/wgpunative_native` with the `.syso` extension so that
the Go linker can include it in a Windows amd64 executable.

## Features

| Feature | Description |
|---|---|
| FFI-compatible API | Reuses the WGPU package sources and API shape from the FFI-oriented Go binding |
| Native call path | Uses Hike/SXSO-generated wrappers rather than GoFFI |
| Static native archive | Links the Rust `wgpu-native` archive through a Go `.syso` object |
| Hike interface | Provides native declarations in `internal/wgpunative/wgpunative.go.hike` |
| Pinned native version | Uses `wgpu-native` tag `v29.0.1.1` |
| Windows triangle example | Includes a native Win32 event loop and a WGSL triangle renderer |

## Requirements

- Go 1.25 or later
- Rust stable toolchain
- `x86_64-pc-windows-gnullvm` Rust target
- MSYS2 UCRT64 clang and lld on Windows
- Hike compiler when regenerating the Hike interface files

The current native archive build is configured for Windows amd64 and the GNU
Windows target. The Rust compiler and Cargo must be selected explicitly when
the MSYS2 installation appears earlier in `PATH` than the Rust installation.

## Build the native library

From `third_party/wgpu-native`, use the following PowerShell configuration:

```powershell
$env:RUSTUP_TOOLCHAIN = "stable"
$env:RUSTC = "$env:USERPROFILE\.cargo\bin\rustc.exe"
$env:LIBCLANG_PATH = "C:\local\gnu\clang+llvm-18.1.8-x86_64-pc-windows-msvc\bin"
$env:CARGO_TARGET_X86_64_PC_WINDOWS_GNULLVM_LINKER = `
  "C:\local\gnu\msys64\ucrt64\bin\clang.exe"
$env:RUSTFLAGS = "-C link-arg=-fuse-ld=lld"
$env:PATH = "C:\Users\$env:USERNAME\.cargo\bin;C:\local\gnu\clang+llvm-18.1.8-x86_64-pc-windows-msvc\bin;" + $env:PATH

rustup target add --toolchain stable x86_64-pc-windows-gnullvm
rustup run stable cargo clean
rustup run stable cargo build `
  --release `
  --target x86_64-pc-windows-gnullvm `
  -j 1
```

The static archive is produced at:

```text
third_party/wgpu-native/target/x86_64-pc-windows-gnullvm/release/libwgpu_native.a
```

Copy the archive to `internal/wgpunative_native` and use the `.syso` extension
expected by the Go linker:

```powershell
Copy-Item `
  third_party/wgpu-native/target/x86_64-pc-windows-gnullvm/release/libwgpu_native.a `
  internal/wgpunative_native/wgpunative_native_windows_amd64.syso
```

## Examples

### Simple link check

```powershell
cd examples\simple
$env:CGO_ENABLED = "1"
$env:CC = "C:\local\gnu\msys64\ucrt64\bin\clang.exe"
go run .
```

### Triangle

The triangle example is a Go port of
`third_party/wgpu-native/examples/triangle`. It creates a native Win32 window,
selects an adapter, creates a WGSL render pipeline, runs the Windows event
loop, and presents a red triangle.

```powershell
cd examples\triangle
$env:CGO_ENABLED = "1"
$env:CC = "C:\local\gnu\msys64\ucrt64\bin\clang.exe"
go run .
```

The example uses the repository's statically linked native archive through the
`.syso` carrier; it does not load `wgpu_native.dll` at runtime.

## Static linking status

The Rust archive and the `.syso` carrier are part of the static-link path. The
Hike declarations in `internal/wgpunative` provide the direct native symbols
needed by the Hike/SXSO integration.

The public `wgpu` package preserves the original API shape, but its static
native entry points are resolved by `internal/wgpunative`; no runtime DLL
loading or GoFFI call layer is used.

## Repository layout

```text
third_party/wgpu-native/        Pinned Rust WGPU native submodule
wgpu/                           Go WebGPU binding based on the FFI project
internal/wgpunative/            Hike native interface declarations
internal/wgpunative_native/     Static archive carrier and link flags
examples/simple/                Static-link smoke test
examples/triangle/              Win32 WGPU triangle example
```

## License

WGPUStatic is released under the MIT License. See [LICENSE](LICENSE).

The project also incorporates or adapts material from the following projects;
their respective licenses and notices remain applicable:

- [`wgpu-native`](third_party/wgpu-native)
- [`webgpu-headers`](third_party/wgpu-native/ffi/webgpu-headers)
