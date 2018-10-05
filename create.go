package main

import (
	"os"
	"fmt"
	"path/filepath"

	"github.com/urfave/cli"
)

var createCommand = cli.Command{
	Name:  "create",
	Usage: "create a container",
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
	spec, err :=  loadSpec(ocffile)

	if err != nil {
		return fmt.Errorf("load config failed: %v", err)
	}
	if container == "" {
		return fmt.Errorf("no container id provided")
	}
/*
	if err = checkConsole(context, spec.Process, attach); err != nil {
		return err
	}
*/

	err = createContainer(container, bundle, root, spec)
	if err != nil {
		return fmt.Errorf("failed to create container: %v", err)
	}

	// write pid file
	if pidf := context.String("pid-file"); pidf != "" {
		f, err := os.OpenFile(pidf,
			os.O_RDWR|os.O_CREATE|os.O_EXCL|os.O_SYNC, 0666)
		if err != nil {
			fmt.Printf("ERR: %s\n", err)
			return err
		}
		// os.Getpid() makes `runu kill` after create.
		// checking pid	by containerd?
		_, err = fmt.Fprintf(f, "%d", os.Getpid() + 1)
		f.Close()
	}

	saveState("created", container, context)

	return nil
}
