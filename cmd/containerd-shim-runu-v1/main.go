//go:build darwin
// +build darwin

/*
   Copyright The containerd Authors.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package main

import (
	"github.com/containerd/containerd/runtime/v2/shim"
)

const (
	// RunuRoot is root directory for runtime execution
	RunuRoot = "/var/run/containerd/runu"
)

func main() {
	shim.Run("io.containerd.runu.v1", New, func(cfg *shim.Config) {
		cfg.NoSetupLogger = false
		// We have own reaper implementation in shim
		cfg.NoSubreaper = true
		cfg.NoReaper = true
	})
}
