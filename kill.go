package main

import (
	"os"
	"fmt"
	"io/ioutil"
	"strconv"
	"syscall"
	"path/filepath"
	"github.com/urfave/cli"
)

var killCommand = cli.Command{
	Name:  "kill",
	ArgsUsage: `<container-id>`,
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  "all",
		},
	},
	Action: func(context *cli.Context) error {
                args := context.Args()
                if args.Present() == false {
                        return fmt.Errorf("Missing container ID")
                }

		root := context.GlobalString("root")
		name := context.Args().Get(0)
		signal, _ := strconv.Atoi(context.Args().Get(1))

		pidFile := filepath.Join(root, name, "runu.pid")
		pid, _ := ioutil.ReadFile(pidFile)
		pid_i, _ := strconv.Atoi(string(pid))

		proc, err := os.FindProcess(pid_i)
		if err != nil {
			return fmt.Errorf("couldn't find pid %d\n", pid_i)
		}
		proc.Signal(syscall.Signal(signal))
		saveState("stopped", name, context)
		return nil
	},
}
