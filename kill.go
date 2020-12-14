package main

import (
	"fmt"

	"github.com/opencontainers/runtime-spec/specs-go"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"os"
	"strconv"
	"syscall"
)

func killFromPidFile(context *cli.Context, pidFile string, signal int) error {
	pid, err := readPidFile(context, pidFile)
	if err != nil {
		// logrus.Warnf("couldn't find pid %d(%s)", pidI, err)
		return nil
	}

	proc, err := os.FindProcess(pid)
	if err != nil {
		logrus.Infof("couldn't find pid %d(%s)", pid, err)
		return err
	}

	err = proc.Signal(syscall.Signal(signal))
	if err != nil {
		logrus.Warnf("couldn't signal to pid %d(%s)", pid, err)
		return err
	}

	return nil
}

var killCommand = cli.Command{
	Name:      "kill",
	ArgsUsage: `<container-id>`,
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name: "all",
		},
	},
	Action: func(context *cli.Context) error {
		args := context.Args()
		if args.Present() == false {
			return fmt.Errorf("Missing container ID")
		}

		container := context.Args().Get(0)
		signal, _ := strconv.Atoi(context.Args().Get(1))

		// kill 9pfs server
		err := killFromPidFile(context, pidFile9p, signal)
		if err != nil {
			logrus.Warnf("killing 9pfs error %s", err)
		}

		// kill main process
		err = killFromPidFile(context, pidFilePriv, signal)
		if err != nil {
			return fmt.Errorf("killing main process error %s", err)
		}

		saveState(specs.StateStopped, container, context)
		return nil
	},
}
