package main

import (
	"fmt"
	"path/filepath"

	"github.com/urfave/cli"
)

var createCommand = cli.Command{
	Name:      "create",
	Usage:     "create a container",
	ArgsUsage: `<container-id>`,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "bundle, b",
			Value: "",
			Usage: `path to the root of the bundle directory, defaults to the current directory`,
		},
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

		cmdCreateUkon(context, false)
		return nil
	},
}

func cmdCreateUkon(context *cli.Context, attach bool) error {
	root := context.GlobalString("root")
	bundle := context.String("bundle")
	container := context.Args().First()
	ocffile := filepath.Join(bundle, specConfig)
	spec, err := loadSpec(ocffile)

	if err != nil {
		return fmt.Errorf("load config failed: %v", err)
	}
	if container == "" {
		return fmt.Errorf("no container id provided")
	}

	err = createContainer(container, bundle, root, spec)
	if err != nil {
		return fmt.Errorf("failed to create container: %v", err)
	}

	prepareUkontainer(context)
	saveState("created", container, context)

	return nil
}
