package main

import (
	"os"
	"path/filepath"
	"github.com/urfave/cli"
)

var deleteCommand = cli.Command{
	Name:  "delete",
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  "force, f",
			Usage: "Forcibly deletes the container if it is still running",
		},
	},
	Action: func(context *cli.Context) error {
		stateFile := filepath.Join("./", "", stateJSON)
		os.Remove(stateFile)
		return nil
	},
}
