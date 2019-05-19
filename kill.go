package main

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"os"
	"strconv"
	"syscall"
)

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
		pidI, err := readPidFile(context, pidFile9p)
		if err != nil {
			logrus.Warnf("couldn't find 9pfs server pid %d(%s)", pidI, err)
		}
		proc, err := os.FindProcess(pidI)
		if err != nil {
			logrus.Warnf("couldn't find 9pfs server pid %d(%s)", pidI, err)
		}
		err = proc.Signal(syscall.SIGTERM)
		logrus.Warnf("killing 9pfs %d %s", syscall.SIGTERM, err)
		if err != nil {
			logrus.Warnf("killing 9pfs error %s", err)
		}
		err = proc.Signal(syscall.Signal(signal))
		logrus.Warnf("killing 9pfs %d %s", signal, err)
		if err != nil {
			logrus.Warnf("killing 9pfs error %s", err)
		}

		// kill main process
		pidI, err = readPidFile(context, pidFilePriv)
		if err != nil {
			logrus.Warnf("couldn't find pid %d(%s)", pidI, err)
		}
		proc, err = os.FindProcess(pidI)
		if err != nil {
			logrus.Warnf("couldn't find pid %d(%s)", pidI, err)
		}
		err = proc.Signal(syscall.Signal(signal))
		if err != nil {
			logrus.Warnf("couldn't signal to pid %d(%s)",
				pidI, err)
		}

		saveState("stopped", container, context)
		return nil
	},
}
