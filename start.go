package main

import (
	"fmt"
	_"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

var startCommand = cli.Command{
	Name:  "start",
	ArgsUsage: `<container-id>`,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "pid-file",
			Usage: "specify the file to write the process id to",
		},
	},
	Action: func(context *cli.Context) error {
                args := context.Args()
                if args.Present() == false {
                        return fmt.Errorf("Missing container ID")
                }

		name := context.Args().First()

		startUnikernel(context)
		saveState("running", name, context)
		// XXX: may need to trigger waited process created
		// by 'create' command.  use libcontainer ???
		return nil
	},
}


