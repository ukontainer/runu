package main

import (
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

var execCommand = cli.Command{
	Name:      "exec",
	Usage:     "execute new process inside the container",
	ArgsUsage: `<container-id> <command> [command options]  || -p process.json <container-id>`,
	Flags:     []cli.Flag{},
	Action: func(context *cli.Context) error {
		logrus.Debug("exec called\n")
		return nil
	},
	SkipArgReorder: true,
}
