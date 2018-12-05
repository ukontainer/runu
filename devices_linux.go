// build +linux

package main

// Device information for linux
var (
	DefaultDevices = []*Device{
		// /dev/urandom
		{
			Path:        "/dev/urandom",
			Type:        'c',
			Major:       1,
			Minor:       9,
			Permissions: "rwm",
			FileMode:    0666,
		},
	}
)
