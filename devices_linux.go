// build +linux

package main

import (
	"C"
	"os"
	"strings"
	"syscall"
	"unsafe"

	"github.com/sirupsen/logrus"
	"golang.org/x/sys/unix"
)

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
		// /dev/net/tun
		{
			Path:        "/dev/net/tun",
			Type:        'c',
			Major:       10,
			Minor:       200,
			Permissions: "rwm",
			FileMode:    0666,
		},
	}
)

const (
	tunDevice        = "/dev/net/tun"
	tunFCsum         = 0x01
	tunFTso4         = 0x02
	virtioNetHdrSize = 12
)

type ifReq struct {
	Name  [syscall.IFNAMSIZ]byte
	Flags uint16
}

func openNetFd(ifname string, specEnv []string) (*os.File, bool) {
	var ifr ifReq
	var vnetHdrSz int
	var offload string

	for _, v := range specEnv {
		if strings.HasPrefix(v, "LKL_OFFLOAD=") {
			offload = "1"
			break
		}
	}

	tapDev, err := os.OpenFile(tunDevice, os.O_RDWR|unix.O_NONBLOCK, 0666)
	if err != nil {
		logrus.Errorf("open %s error: %s\n", tunDevice, err)
		panic(err)
	}

	copy(ifr.Name[:(syscall.IFNAMSIZ-1)], ifname)
	ifr.Flags = syscall.IFF_TAP | syscall.IFF_NO_PI

	if offload != "" {
		ifr.Flags |= syscall.IFF_VNET_HDR
		vnetHdrSz = virtioNetHdrSize
	}

	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL,
		uintptr(tapDev.Fd()),
		uintptr(syscall.TUNSETIFF),
		uintptr(unsafe.Pointer(&ifr)),
	)
	if errno != 0 {
		panic(errno)
	}

	if offload != "" {
		/* XXX: offload feature should be configurable */
		_, _, errno := syscall.Syscall(syscall.SYS_IOCTL,
			uintptr(tapDev.Fd()),
			uintptr(syscall.TUNSETVNETHDRSZ),
			uintptr(unsafe.Pointer(&vnetHdrSz)),
		)
		if errno != 0 {
			panic(errno)
		}

		tapArg := tunFCsum | tunFTso4
		_, _, errno = syscall.Syscall(syscall.SYS_IOCTL,
			uintptr(tapDev.Fd()),
			uintptr(syscall.TUNSETOFFLOAD),
			uintptr(tapArg),
		)
		if errno != 0 {
			panic(errno)
		}
	}

	return tapDev, true
}
