// build +darwin

package main

import (
	"C"
	"os"
	_ "syscall"
	_ "unsafe"

	"github.com/sirupsen/logrus"
	"golang.org/x/sys/unix"
)

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

func openNetFd(ifname string, specEnv []string) (*os.File, bool) {
	tapDev, err := os.OpenFile("/dev/"+ifname, os.O_RDWR | unix.O_NONBLOCK, 0666)
	if err != nil {
		logrus.Errorf("open %s error: /dev/%s\n", ifname, err)
		panic(err)
	}

	/* XXX: no way to non-block ? */
	/*
	on := 1

	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL,
		uintptr(tapDev.Fd()),
		uintptr(unix.FIONBIO),
		uintptr(unsafe.Pointer(&on)),
	)
	if errno != 0 {
		panic(errno)
	}
	*/


	return tapDev, true
}
