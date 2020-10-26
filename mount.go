package main

import (
	"fmt"
	"strings"

	"github.com/opencontainers/runtime-spec/specs-go"
)

func checkFsFlags(spec *specs.Spec) (bool, bool, error) {
	use9pFs := false
	hasRootFs := false

	for _, env := range spec.Process.Env {
		if strings.HasPrefix(env, "LKL_USE_9PFS=") {
			use9pFs = true
		}
		if strings.HasPrefix(env, "LKL_ROOTFS=") {
			hasRootFs = true
		}
	}
	if hasRootFs && use9pFs {
		return false, false, fmt.Errorf("LKL_ROOTFS and LKL_USE_9PFS cannot be specified at the same time")
	}
	return use9pFs, hasRootFs, nil
}
