package main

import (
	"context"
	"io"
	"net"
	"os"

	p9p "github.com/docker/go-p9p"
	"github.com/docker/go-p9p/ufs"
	"github.com/sirupsen/logrus"
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
	ctx := context.Background()
	l, err := net.Listen("tcp", addr9p)
	if err != nil {
		panic(err)
	}
	defer l.Close()

	for {
		c, err := l.Accept()
		if err != nil {
			panic(err)
		}

		go func(conn net.Conn) {
			ctx := context.WithValue(ctx, contextKey("conn"), conn)
			session, err := ufs.NewSession(ctx, path)
			if err != nil {
				logrus.Println("error creating session", err)
				return
			}

			if err := p9p.ServeConn(ctx, conn, p9p.Dispatch(session)); err != nil && err != io.EOF {
				logrus.Println("error serving conn:", err)
			}
		}(c)
	}
}
