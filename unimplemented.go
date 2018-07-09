package main

import (
	"os"
	"fmt"
	"path/filepath"
	"github.com/urfave/cli"
)

var startCommand = cli.Command{
	Name:  "start",
	Action: func(context *cli.Context) error {
		saveState("created", context)
		return cmdStartUnikernel(context)
	},
}

var killCommand = cli.Command{
	Name:  "kill",
	Action: func(context *cli.Context) error {
		saveState("stopped", context)
		panic(nil)
		return nil
	},
}

var deleteCommand = cli.Command{
	Name:  "delete",
	Action: func(context *cli.Context) error {
		spec, err := setupSpec(context)
		if err != nil {
			fmt.Printf("setupSepc err\n")
			return err
		}

		rootfs,_ := filepath.Abs(spec.Root.Path)
		stateFile := filepath.Join(rootfs, "", stateJSON)
		os.Remove(stateFile)
		return nil
	},
}

