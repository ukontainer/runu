package main

import (
	"fmt"
	"io/ioutil"
	"strconv"
	"os"
	"syscall"
	"path/filepath"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

var startCommand = cli.Command{
	Name:  "start",
	ArgsUsage: `<container-id>`,
	Flags: []cli.Flag{
	},
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
	root := context.GlobalString("root")
	pidFile := filepath.Join(root, container, pid_file_priv)
	pid, _ := ioutil.ReadFile(pidFile)
	pid_i, _ := strconv.Atoi(string(pid))

	proc, err := os.FindProcess(pid_i)
	if err != nil {
		return fmt.Errorf("couldn't find pid %d\n", pid_i)
	}

	// wake the process
	logrus.Debugf("proc %s, pid=%d", proc, pid_i)
	proc.Signal(syscall.Signal(syscall.SIGCONT))

	// catch child errors if possible
	go func() {
		proc_stat, _ := proc.Wait()
		if proc_stat != nil {
			waitstatus := proc_stat.Sys().(syscall.WaitStatus)
			if waitstatus.Signal() != syscall.SIGINT &&
				waitstatus.Signal() != syscall.SIGTERM &&
				!waitstatus.Exited() {
				panic(proc_stat)
			}
		}
		if proc_stat == nil {
			logrus.Debugf("no process to wait. non-child process? (%d)",
				pid_i)
		}
	}()

	return nil
}
