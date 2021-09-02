// +build !windows

/*
   Copyright The containerd Authors.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package main

import (
	"context"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"sync"
	"time"

	"github.com/containerd/containerd/pkg/process"
	runc "github.com/containerd/go-runc"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/sys/unix"
)

// ErrNoSuchProcess is returned when the process no longer exists
var ErrNoSuchProcess = errors.New("no such process")

const bufferSize = 32

type subscriber struct {
	sync.Mutex
	c      chan runc.Exit
	closed bool
}

func (s *subscriber) close() {
	s.Lock()
	if s.closed {
		s.Unlock()
		return
	}
	close(s.c)
	s.closed = true
	s.Unlock()
}

func (s *subscriber) do(fn func()) {
	s.Lock()
	fn()
	s.Unlock()
}

// Reap should be called when the process receives an SIGCHLD.  Reap will reap
// all exited processes and close their wait channels
func Reap() error {
	now := time.Now()
	exits, err := reap(false)
	if err != nil {
		return err
	}
	exits, err = reapOS(exits)
	for _, e := range exits {
		done := Default.notify(runc.Exit{
			Timestamp: now,
			Pid:       e.Pid,
			Status:    e.Status,
		})

		select {
		case <-done:
		case <-time.After(1 * time.Second):
		}
	}
	return err
}

// Default is the default monitor initialized for the package
var Default = &Monitor{
	subscribers: make(map[chan runc.Exit]*subscriber),
}

// Monitor monitors the underlying system for process status changes
type Monitor struct {
	sync.Mutex

	subscribers map[chan runc.Exit]*subscriber
}

// Start starts the command a registers the process with the reaper
func (m *Monitor) Start(c *exec.Cmd) (chan runc.Exit, error) {
	ec := m.Subscribe()
	if err := c.Start(); err != nil {
		m.Unsubscribe(ec)
		return nil, err
	}
	return ec, nil
}

// Wait blocks until a process is signal as dead.
// User should rely on the value of the exit status to determine if the
// command was successful or not.
func (m *Monitor) Wait(c *exec.Cmd, ec chan runc.Exit) (int, error) {
	for e := range ec {
		if e.Pid == c.Process.Pid {
			// make sure we flush all IO
			c.Wait()
			m.Unsubscribe(ec)
			return e.Status, nil
		}
	}
	// return no such process if the ec channel is closed and no more exit
	// events will be sent
	return -1, ErrNoSuchProcess
}

// Subscribe to process exit changes
func (m *Monitor) Subscribe() chan runc.Exit {
	c := make(chan runc.Exit, bufferSize)
	m.Lock()
	m.subscribers[c] = &subscriber{
		c: c,
	}
	m.Unlock()
	return c
}

// Unsubscribe to process exit changes
func (m *Monitor) Unsubscribe(c chan runc.Exit) {
	m.Lock()
	s, ok := m.subscribers[c]
	if !ok {
		m.Unlock()
		return
	}
	s.close()
	delete(m.subscribers, c)
	m.Unlock()
}

func (m *Monitor) getSubscribers() map[chan runc.Exit]*subscriber {
	out := make(map[chan runc.Exit]*subscriber)
	m.Lock()
	for k, v := range m.subscribers {
		out[k] = v
	}
	m.Unlock()
	return out
}

func (m *Monitor) notify(e runc.Exit) chan struct{} {
	const timeout = 1 * time.Millisecond
	var (
		done    = make(chan struct{}, 1)
		timer   = time.NewTimer(timeout)
		success = make(map[chan runc.Exit]struct{})
	)
	stop(timer, true)

	go func() {
		defer close(done)

		for {
			var (
				failed      int
				subscribers = m.getSubscribers()
			)
			for _, s := range subscribers {
				s.do(func() {
					if s.closed {
						return
					}
					if _, ok := success[s.c]; ok {
						return
					}
					timer.Reset(timeout)
					recv := true
					select {
					case s.c <- e:
						success[s.c] = struct{}{}
					case <-timer.C:
						recv = false
						failed++
					}
					stop(timer, recv)
				})
			}
			// all subscribers received the message
			if failed == 0 {
				return
			}
		}
	}()
	return done
}

func stop(timer *time.Timer, recv bool) {
	if !timer.Stop() && recv {
		<-timer.C
	}
}

// exit is the wait4 information from an exited process
type exit struct {
	Pid    int
	Status int
}

// reap reaps all child processes for the calling process and returns their
// exit information
func reap(wait bool) (exits []exit, err error) {
	var (
		ws  unix.WaitStatus
		rus unix.Rusage
	)
	flag := unix.WNOHANG
	if wait {
		flag = 0
	}
	for {
		pid, err := unix.Wait4(-1, &ws, flag, &rus)
		if err != nil {
			if err == unix.ECHILD {
				return exits, nil
			}
			return exits, err
		}
		if pid <= 0 {
			return exits, nil
		}
		exits = append(exits, exit{
			Pid:    pid,
			Status: exitStatus(ws),
		})
	}
}

// reapOS is additional reap process upon receipt of SIGCHLD.
// Since macOS doesn't raise SIGCHLD on orphaned children's exit,
// reapOS polls the status of registered process and terminate it
// if it's already exited.
func reapOS(exits []exit) ([]exit, error) {
	pid, err := runc.ReadPidFile(filepath.Join("", process.InitPidFile))
	if pid <= 0 {
		return exits, errors.Errorf("can't find pid=%d %s", pid, err)
	}

	process, err := os.FindProcess(pid)
	// ensure the process is running
	if process != nil {
		// from kill(2):
		// A value of 0, however, will cause error checking to be
		// performed (with no signal being sent).
		// This can be used to check the validity of pid.
		err = process.Signal(syscall.Signal(0))
	}
	logrus.Debugf("checking pid=%d proc=%v err=%v", pid, process, err)

	// if process exists && already finished
	if err != nil && strings.Contains(err.Error(), "os: process already finished") {
		exits = append(exits, exit{
			Pid:    pid,
			Status: 0, // XXX
		})

		logrus.Debugf("reapOS: detect exited, pid=%d", pid)
	}

	return exits, nil
}

const exitSignalOffset = 128

// exitStatus returns the correct exit status for a process based on if it
// was signaled or exited cleanly
func exitStatus(status unix.WaitStatus) int {
	if status.Signaled() {
		return exitSignalOffset + int(status.Signal())
	}
	return status.ExitStatus()
}

func SetupReaperSignals(ctx context.Context, logger *logrus.Entry) error {
	signals := make(chan os.Signal, 32)
	signal.Notify(signals, unix.SIGCHLD)
	go handleSignals(ctx, logger, signals)
	return nil
}

// copied from containerd code
func handleSignals(ctx context.Context, logger *logrus.Entry, signals chan os.Signal) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case s := <-signals:
			switch s {
			case unix.SIGCHLD:
				if err := Reap(); err != nil {
					logger.WithError(err).Error("reap exit status")
				}
			case unix.SIGPIPE:
			}
		}
	}
}
