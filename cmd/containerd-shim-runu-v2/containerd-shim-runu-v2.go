package main

import (
	"github.com/containerd/containerd/runtime/v2/shim"
)

const (
	// RuntimeV2 is the name of runtime
	RuntimeV2 = "io.containerd.runu.v2"
)

func main() {
	shim.Run("io.containerd.runu.v2", New)
}
