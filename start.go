package main

import (
	"github.com/urfave/cli"
)

var startCommand = cli.Command{
	Name:  "start",
	Action: func(context *cli.Context) error {
		saveState("created", -1, context)
		return cmdStartUnikernel(context)
	},
}


