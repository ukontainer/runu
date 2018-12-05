// build +darwin

package main

// Device information for macOS
var (
	DefaultDevices = []*Device{
		// /dev/urandom
		{
			Path:        "/dev/urandom",
			Type:        'c',
			Major:       14,
			Minor:       1,
			Permissions: "rwm",
			FileMode:    0666,
		},
	}
)
