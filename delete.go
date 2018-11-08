package main

import (
	"fmt"
	"github.com/urfave/cli"
)

var deleteCommand = cli.Command{
	Name:      "delete",
	ArgsUsage: `<container-id>`,
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  "force, f",
			Usage: "Forcibly deletes the container if it is still running",
		},
	},
	Action: func(context *cli.Context) error {
		args := context.Args()
		if args.Present() == false {
			return fmt.Errorf("Missing container ID")
		}

		container := context.Args().First()
		return deleteContainer(context.GlobalString("root"), container)
	},
}
