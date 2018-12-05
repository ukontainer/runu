package main

import (
	"fmt"

	"github.com/urfave/cli"
)

// default action is to start a container
var runCommand = cli.Command{
	Name:      "run",
	ArgsUsage: `<container-id>`,
	Usage:     "create and run a container",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "bundle, b",
			Value: "",
			Usage: `path to the root of the bundle directory, defaults to the current directory`,
		},
	},
	Action: func(context *cli.Context) error {
		args := context.Args()
		if args.Present() == false {
			return fmt.Errorf("Missing container ID")
		}

		// XXX: create + start
		container := context.Args().Get(0)
		err := cmdCreateUkon(context, false)
		if err != nil {
			return err
		}
		return resumeUkontainer(context, container)

	},
}
