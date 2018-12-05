package main

import (
	"fmt"
	"github.com/urfave/cli"
	"io/ioutil"
	"os"
	"path/filepath"
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

		root := context.GlobalString("root")
		container := context.Args().Get(0)
		signal, _ := strconv.Atoi(context.Args().Get(1))

		pidFile := filepath.Join(root, container, pidFilePriv)
		pid, err := ioutil.ReadFile(pidFile)
		if err != nil {
			return fmt.Errorf("couldn't find pidfile %s(%s)",
				pidFilePriv, err)
		}
		pidI, _ := strconv.Atoi(string(pid))

		proc, err := os.FindProcess(pidI)
		if err != nil {
			return fmt.Errorf("couldn't find pid %d(%s)", pidI, err)
		}
		err = proc.Signal(syscall.Signal(signal))
		if err != nil {
			return fmt.Errorf("couldn't signal to pid %d(%s)",
				pidI, err)
		}
		saveState("stopped", container, context)
		return nil
	},
}
