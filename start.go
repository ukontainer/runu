package main

import (
	"fmt"

	"github.com/opencontainers/runtime-spec/specs-go"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"os"
	"syscall"
)

var startCommand = cli.Command{
	Name:      "start",
	ArgsUsage: `<container-id>`,
	Flags:     []cli.Flag{},
	Action: func(context *cli.Context) error {
		args := context.Args()
		if args.Present() == false {
			return fmt.Errorf("Missing container ID")
		}

		container := context.Args().Get(0)
		resumeUkontainer(context, container)
		saveState(specs.StateRunning, container, context)
		return nil
	},
}

func resumeUkontainer(context *cli.Context, container string) error {
	// wake the process
	pidI, err := readPidFile(context, pidFilePriv)
	if err != nil {
		return fmt.Errorf("couldn't find pid %d", pidI)
	}
	proc, err := os.FindProcess(pidI)
	if err != nil {
		return fmt.Errorf("couldn't find pid %d", pidI)
	}

	logrus.Debugf("proc %p, pid=%d", proc, pidI)
	proc.Signal(syscall.Signal(syscall.SIGCONT))

	return nil
}
