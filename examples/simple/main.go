package main

import (
	"fmt"

	"github.com/kanryu/wgpu-static/internal/wgpunative"
	_ "github.com/kanryu/wgpu-static/internal/wgpunative_native"
)

func main() {
	version := wgpunative.GetVersion()
	fmt.Printf("wgpu-native version: %d\n", version)
}
