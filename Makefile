SHELL := powershell.exe
.SHELLFLAGS := -NoProfile -ExecutionPolicy Bypass -Command

TARGET := x86_64-pc-windows-gnullvm
WGPU_DIR := third_party/wgpu-native
WGPU_MANIFEST := $(WGPU_DIR)/Cargo.toml
WGPU_TARGET_DIR := $(WGPU_DIR)/target/$(TARGET)
ARCHIVE := $(WGPU_TARGET_DIR)/release/libwgpu_native.a
DEBUG_ARCHIVE := $(WGPU_TARGET_DIR)/debug/libwgpu_native.a
SYSO := internal/wgpunative_native/wgpunative_native_windows_amd64.syso

RUSTC ?= $(USERPROFILE)/.cargo/bin/rustc.exe
LINKER ?= C:/local/gnu/msys64/ucrt64/bin/clang.exe
LIBCLANG_PATH ?= C:/local/gnu/clang+llvm-18.1.8-x86_64-pc-windows-msvc/bin

.PHONY: all native native-release native-debug native-clean clean

all: native-release

native: native-release

native-release:
	$$env:RUSTUP_TOOLCHAIN = 'stable'; $$env:RUSTC = '$(RUSTC)'; $$env:LIBCLANG_PATH = '$(LIBCLANG_PATH)'; $$env:PATH = '$(LIBCLANG_PATH);' + $$env:PATH; $$env:CARGO_TARGET_X86_64_PC_WINDOWS_GNULLVM_LINKER = '$(LINKER)'; $$env:RUSTFLAGS = '-C link-arg=-fuse-ld=lld'; & rustup target add --toolchain stable $(TARGET); if ($$LASTEXITCODE -ne 0) { exit $$LASTEXITCODE }; & rustup run stable cargo build --manifest-path '$(WGPU_MANIFEST)' --release --target $(TARGET) -j 1; if ($$LASTEXITCODE -ne 0) { exit $$LASTEXITCODE }; Copy-Item -LiteralPath '$(ARCHIVE)' -Destination '$(SYSO)' -Force

native-debug:
	$$env:RUSTUP_TOOLCHAIN = 'stable'; $$env:RUSTC = '$(RUSTC)'; $$env:LIBCLANG_PATH = '$(LIBCLANG_PATH)'; $$env:PATH = '$(LIBCLANG_PATH);' + $$env:PATH; $$env:CARGO_TARGET_X86_64_PC_WINDOWS_GNULLVM_LINKER = '$(LINKER)'; $$env:RUSTFLAGS = '-C link-arg=-fuse-ld=lld'; & rustup target add --toolchain stable $(TARGET); if ($$LASTEXITCODE -ne 0) { exit $$LASTEXITCODE }; & rustup run stable cargo build --manifest-path '$(WGPU_MANIFEST)' --target $(TARGET) -j 1; if ($$LASTEXITCODE -ne 0) { exit $$LASTEXITCODE }; Copy-Item -LiteralPath '$(DEBUG_ARCHIVE)' -Destination '$(SYSO)' -Force

native-clean:
	& rustup run stable cargo clean --manifest-path '$(WGPU_MANIFEST)'

clean: native-clean
