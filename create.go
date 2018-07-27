package main

import (
	"os"
	"fmt"
	"github.com/urfave/cli"
)

var createCommand = cli.Command{
	Name:  "create",
	Usage: "create a container",
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
		if pidf := context.String("pid-file"); pidf != "" {
			f, err := os.OpenFile(pidf, os.O_RDWR|os.O_CREATE|os.O_EXCL|os.O_SYNC, 0666)
			if err != nil {
				fmt.Printf("ERR: %s\n", err)
				return err
			}
			_, err = fmt.Fprintf(f, "%d", os.Getppid())
			f.Close()
		}
		saveState("created", -1, context)

		return nil
	},
}