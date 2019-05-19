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
	TUN_DEVICE = "/dev/net/tun"
	TUN_F_CSUM = 0x01
	TUN_F_TSO4 = 0x02
	TUN_F_TSO6 = 0x04
	TUN_F_TSO_ECN =0x08
	TUN_F_UFO = 0x10
	virtio_net_hdr_size = 12
)

type ifReq struct {
	Name  [syscall.IFNAMSIZ]byte
	Flags uint16
}

func openNetFd(ifname string, specEnv []string) (*os.File, bool) {
	var ifr ifReq
	var vnet_hdr_sz int
	var offload string

	for _, v := range specEnv {
		if strings.HasPrefix(v, "LKL_OFFLOAD=") {
			offload = "1"
			break
		}
	}

	tapDev, err := os.OpenFile(TUN_DEVICE, os.O_RDWR | unix.O_NONBLOCK, 0666)
	if err != nil {
		logrus.Errorf("open %s error: %s\n", TUN_DEVICE, err)
		panic(err)
	}

	copy(ifr.Name[:(syscall.IFNAMSIZ-1)], ifname)
	ifr.Flags = syscall.IFF_TAP | syscall.IFF_NO_PI

	if offload != "" {
		ifr.Flags |= syscall.IFF_VNET_HDR
		vnet_hdr_sz = virtio_net_hdr_size
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
			uintptr(unsafe.Pointer(&vnet_hdr_sz)),
		)
		if errno != 0 {
			panic(errno)
		}

		tap_arg := TUN_F_CSUM | TUN_F_TSO4
		_, _, errno = syscall.Syscall(syscall.SYS_IOCTL,
			uintptr(tapDev.Fd()),
			uintptr(syscall.TUNSETOFFLOAD),
			uintptr(tap_arg),
		)
		if errno != 0 {
			panic(errno)
		}
	}

	return tapDev, true
}
