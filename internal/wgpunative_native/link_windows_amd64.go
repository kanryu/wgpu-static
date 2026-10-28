//go:build windows && amd64

package wgpunative_native

// #cgo windows,amd64 LDFLAGS: -lntdll -ladvapi32 -luser32 -luserenv -lkernel32 -lole32 -loleaut32 -lshell32 -lws2_32 -lucrt -lucrtbase -lm -lgcc -lgcc_eh -lmingwex -lmingw32
import "C"

// wgpu-native is built for the Windows GNU target. Its archive leaves the
// Windows and MinGW runtime libraries to the final external link. This tiny
// cgo-only helper carries those flags; the public Go binding remains usable
// without exposing cgo in its API.
