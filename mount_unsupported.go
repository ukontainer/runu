// +build !linux

package main

import (
	"fmt"
	goruntime "runtime"

	"github.com/opencontainers/runtime-spec/specs-go"
)

func doMounts(spec *specs.Spec) (bool, error) {

	for _, m := range spec.Mounts {
		switch m.Type {
		case "bind":
			doContinue := false
			for _, d := range []string{"/etc/hosts", "/etc/hostname", "/etc/resolv.conf", "/dev/shm"} {
				if m.Destination == d {
					doContinue = true
					break
				}
			}

			if doContinue {
				continue
			}

			return false, fmt.Errorf("volume mount is not supported on %s: %v", goruntime.GOOS, m)
		}
	}
	return false, nil
}
