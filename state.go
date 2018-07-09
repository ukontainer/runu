package main

import (
	"os"
	"fmt"
	"io/ioutil"
	"encoding/json"
	"path/filepath"
	"github.com/urfave/cli"
	"github.com/opencontainers/runtime-spec/specs-go"
)

func saveState(status string, context *cli.Context) error {
	spec, err := setupSpec(context)
	if err != nil {
		fmt.Printf("setupSepc err\n")
		return err
	}

	rootfs,_ := filepath.Abs(spec.Root.Path)
	stateFile := filepath.Join(rootfs, "", stateJSON)
	cs := &specs.State {
		Version: spec.Version,
		ID: "1",
		Status: status,
		Pid: 1,
		Bundle: rootfs,
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
		spec, _ := setupSpec(context)
		rootfs,_ := filepath.Abs(spec.Root.Path)
		stateFile := filepath.Join(rootfs, "", stateJSON)
		stateData, _ := ioutil.ReadFile(stateFile)

		os.Stdout.Write(stateData)
		return nil
	},
}

