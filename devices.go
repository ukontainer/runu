package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"golang.org/x/sys/unix"
)

// Device information
type Device struct {
	// Device type, block, char, etc.
	Type rune `json:"type"`

	// Path to the device.
	Path string `json:"path"`

	// Major is the device's major number.
	Major int64 `json:"major"`

	// Minor is the device's minor number.
	Minor int64 `json:"minor"`

	// Cgroup permissions format, rwm.
	Permissions string `json:"permissions"`

	// FileMode permission bits for the device.
	FileMode os.FileMode `json:"file_mode"`

	// Uid of the device.
	Uid uint32 `json:"uid"`

	// Gid of the device.
	Gid uint32 `json:"gid"`

	// Write the file to the allowed list
	Allow bool `json:"allow"`
}

func createDeviceNode(rootfs string, node *Device) error {
	dest := filepath.Join(rootfs, node.Path)
	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		return err
	}

	return mknodDevice(dest, node)
}

func mknodDevice(dest string, node *Device) error {
	fileMode := node.FileMode
	switch node.Type {
	case 'c', 'u':
		fileMode |= unix.S_IFCHR
	case 'b':
		fileMode |= unix.S_IFBLK
	case 'p':
		fileMode |= unix.S_IFIFO
	default:
		return fmt.Errorf("%c is not a valid device type for device %s", node.Type, node.Path)
	}

	dNum := int((node.Major << 8) | (node.Minor & 0xff) | ((node.Minor & 0xfff00) << 12))
	if err := unix.Mknod(dest, uint32(fileMode), dNum); err != nil {
		return err
	}
	return unix.Chown(dest, int(node.Uid), int(node.Gid))
}

func openRootfsFd(file string) (*os.File, bool) {
	fd, err := os.OpenFile(file, os.O_RDWR, 0666)
	if err != nil {
		logrus.Errorf("open %s error: /dev/%s\n", file, err)
		panic(err)
	}

	return fd, true
}

func openJsonFd(file string) (*os.File, bool) {
	fd, err := os.OpenFile(file, os.O_RDONLY, unix.S_IRUSR | unix.S_IWUSR)
	if err != nil {
		logrus.Errorf("open %s error: %s\n", file, err)
		panic(err)
	}

	return fd, false
}
