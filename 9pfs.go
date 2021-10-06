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
	addr9p    = "127.0.0.1:5640"
	p9Timeout = 30
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
func start9pfsServer(path string) error {
	ctx := context.Background()
	l, err := net.Listen("tcp", addr9p)
	if err != nil {
		panic(err)
	}
	defer l.Close()

	for {
		statusChan := make(chan error)
		c, err := l.Accept()
		if err != nil {
			panic(err)
		}

		go func(conn net.Conn) error {
			ctx := context.WithValue(ctx, contextKey("conn"), conn)
			session, err := ufs.NewSession(ctx, path)
			if err != nil {
				logrus.Println("error creating session", err)
				statusChan <- err
			}

			if err := p9p.ServeConn(ctx, conn, p9p.Dispatch(session), p9Timeout); err != nil && err != io.EOF {
				logrus.Printf("error serving conn: (path=%s) %+v", path, err)
				statusChan <- err
			}
			return nil
		}(c)
		return <-statusChan
	}
}
