package main

import (
	"os"
	"fmt"
	"io/ioutil"
	"encoding/json"
	"path/filepath"
	"github.com/urfave/cli"
	"github.com/sirupsen/logrus"
	"github.com/opencontainers/runtime-spec/specs-go"
)

func saveState(status string, Pid int, context *cli.Context) error {
	spec, err := setupSpec(context)
	if err != nil {
		fmt.Printf("setupSepc err\n")
		return err
	}

	rootfs,_ := filepath.Abs(spec.Root.Path)
	stateFile := filepath.Join("./", "", stateJSON)
	cs := &specs.State {
		Version: spec.Version,
		ID: context.Args().Get(0),
		Status: status,
		Bundle: rootfs,
	}
	if Pid != -1 {
		cs.Pid = Pid
	}
	stateData, _ := json.MarshalIndent(cs, "", "\t")

	if err := ioutil.WriteFile(stateFile, stateData, 0666); err != nil {
		panic(err);
		return err
	}

	return nil
}

var stateCommand = cli.Command{
	Name:  "state",
	Action: func(context *cli.Context) error {
		stateFile := filepath.Join("./", "", stateJSON)
		stateData, _ := ioutil.ReadFile(stateFile)

		os.Stdout.Write(stateData)
		logrus.Printf("stat = %s\n", stateData)
		return nil
	},
}

