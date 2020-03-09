// build +linux

package main

import (
	"C"
	"net"
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

const (
	_ETHTOOL_GTXCSUM = 0x00000016 // linux/ethtool.h
	_ETHTOOL_STXCSUM = 0x00000017 // linux/ethtool.h
)

type ifReq struct {
	Name  [syscall.IFNAMSIZ]byte
	Flags uint16
}

type ifReqData struct {
	Name [syscall.IFNAMSIZ]byte
	Data uintptr
}

// linux/ethtool.h 'struct ethtool_value'
type ethtoolValue struct {
	Cmd  uint32
	Data uint32
}

func openNetTapFd(ifname string, specEnv []string) (*os.File, bool) {
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

func htons(i uint16) uint16 {
	return (i<<8)&0xff00 | i>>8
}

func disableTxCsumOffloadForRawsock(ifname string) error {
	// XXX: disable tx csum offload (may need vnet_hdr impl in raw sock)
	iocSock, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_DGRAM, 0)
	if err != nil {
		logrus.Errorf("socket for ioctl failure error: %s\n", err)
		panic(err)
	}
	defer syscall.Close(iocSock)

	value := ethtoolValue{Cmd: _ETHTOOL_STXCSUM, Data: 0}
	request := ifReqData{Data: uintptr(unsafe.Pointer(&value))}
	copy(request.Name[:], ifname)

	_, _, errno := syscall.RawSyscall(syscall.SYS_IOCTL, uintptr(iocSock),
		uintptr(unix.SIOCETHTOOL), uintptr(unsafe.Pointer(&request)))
	if errno != 0 {
		logrus.Errorf("disabling csum offload failure: %d\n", errno)
		panic("ETHTOOL_STXCSUM")
	}

	value.Cmd = _ETHTOOL_GTXCSUM
	_, _, errno = syscall.RawSyscall(syscall.SYS_IOCTL, uintptr(iocSock),
		uintptr(unix.SIOCETHTOOL), uintptr(unsafe.Pointer(&request)))
	if errno != 0 {
		logrus.Errorf("disabling csum offload failure: %d\n", errno)
		panic("ETHTOOL_GTXCSUM")
	}
	logrus.Debugf("ifreq rx csum flag is %d\n", value.Data)

	return nil
}

func openNetRawsockFd(ifname string, specEnv []string) (*os.File, bool) {
	fd, err := syscall.Socket(syscall.AF_PACKET, syscall.SOCK_RAW,
		int(htons(syscall.ETH_P_ALL)))
	if err != nil {
		logrus.Errorf("open raw socket error: %s\n", err)
		panic(err)
	}

	// bind to ifname
	ifi, err := net.InterfaceByName(ifname)
	if err != nil {
		logrus.Errorf("can't find interface %s: %s\n", ifname, err)
		panic(err)
	}

	sa := &unix.SockaddrLinklayer{
		Protocol: htons(unix.ETH_P_ALL),
		Ifindex:  ifi.Index,
	}
	err = unix.Bind(fd, sa)
	if err != nil {
		logrus.Errorf("can't bind to interface %s: %s\n", ifname, err)
		panic(err)
	}

	// set promisc
	mreq := unix.PacketMreq{
		Ifindex: int32(ifi.Index),
		Type:    unix.PACKET_MR_PROMISC,
	}
	if err = unix.SetsockoptPacketMreq(int(fd), unix.SOL_PACKET,
		unix.PACKET_ADD_MEMBERSHIP, &mreq); err != nil {
		logrus.Errorf("set nonblocking error: %s\n", err)
		panic(err)
	}

	// vnethdr
	var on uint64
	if err = unix.SetsockoptUint64(int(fd), unix.SOL_PACKET,
		unix.PACKET_VNET_HDR, on); err != nil {
		logrus.Errorf("set vnethdr sockopt error: %s\n", err)
		panic(err)
	}

	// set nonblock sock
	if err = unix.SetNonblock(fd, true); err != nil {
		logrus.Errorf("set nonblocking error: %s\n", err)
		panic(err)
	}

	return os.NewFile(uintptr(fd), "eth-packet-socket"), true
}

func openNetFd(ifname string, specEnv []string) (*os.File, bool) {
	if strings.HasPrefix(ifname, "eth") {
		return openNetRawsockFd(ifname, specEnv)
	} else if strings.HasPrefix(ifname, "tap") {
		return openNetTapFd(ifname, specEnv)
	}

	return nil, false
}
