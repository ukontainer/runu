package main

import (
	"net"
	"os"

	"github.com/hugelgupf/p9/fsimpl/localfs"
	"github.com/hugelgupf/p9/p9"
	"github.com/sirupsen/logrus"
	"github.com/u-root/u-root/pkg/ulog"
)

const (
	addr9p = "127.0.0.1:5640"
)

type contextKey string

// 9pfs client side
func connect9pfs() (*os.File, bool) {
	tcpAddr, err := net.ResolveTCPAddr("tcp", addr9p)
	if err != nil {
		panic(err)
	}

	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		logrus.Errorf("open %s error: %s\n", tcpAddr, err)
		panic(err)
	}

	fd, err := conn.File()
	if err != nil {
		logrus.Errorf("File() %s error: %s\n", conn.RemoteAddr(), err)
		panic(err)
	}

	conn.Close()

	// XXX: 9pfs sv requires blocking socket ??
	return fd, false
}

// 9pfs server side
func start9pfsServer(path string) {
	l, err := net.Listen("tcp", addr9p)
	if err != nil {
		panic(err)
	}
	defer l.Close()

	// TODO: currently not in use
	var opts []p9.ServerOpt
	if false {
		opts = append(opts, p9.WithServerLogger(ulog.Log))
	}
	// Run the server.
	s := p9.NewServer(localfs.Attacher(path), opts...)
	s.Serve(l)
}
