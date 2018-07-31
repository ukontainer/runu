package main

import (
       _"os"
       _"github.com/sirupsen/logrus"
       "github.com/urfave/cli"
)

var startCommand = cli.Command{
	Name:  "start",
	Action: func(context *cli.Context) error {
		saveState("running", -1, context)
		// XXX: may need to trigger waited process created
		// by 'create' command.  use libcontainer ???
		return nil
	},
}


