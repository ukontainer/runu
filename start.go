package main

import (
	"fmt"
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
		saveState("running", container, context)
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

	// catch child errors if possible
	go func() {
		procStat, _ := proc.Wait()
		if procStat != nil {
			waitstatus := procStat.Sys().(syscall.WaitStatus)
			if waitstatus.Signal() != syscall.SIGINT &&
				waitstatus.Signal() != syscall.SIGTERM &&
				!waitstatus.Exited() {
				fmt.Printf("err %s\n", err)
				panic(procStat)
			}
		}
		if procStat == nil {
			logrus.Debugf("no process to wait. non-child process? (%d)",
				pidI)
		}
	}()

	return nil
}
