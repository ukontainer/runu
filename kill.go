package main

import (
       "os"
       "github.com/urfave/cli"
)

var killCommand = cli.Command{
	Name:  "kill",
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  "all",
		},
	},
	Action: func(context *cli.Context) error {
		saveState("stopped", -1, context)

		proc, _ := os.FindProcess(os.Getppid())
		proc.Signal(os.Interrupt)
		return nil
	},
}
