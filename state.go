package main

import (
	"encoding/json"
	"fmt"
	"github.com/opencontainers/runtime-spec/specs-go"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"io/ioutil"
	"os"
	"path/filepath"
)

var stateCommand = cli.Command{
	Name:      "state",
	ArgsUsage: `<container-id>`,
	Action: func(context *cli.Context) error {
		args := context.Args()
		if args.Present() == false {
			return fmt.Errorf("Missing container ID")
		}

		root := context.GlobalString("root")
		name := context.Args().First()
		stateFile := filepath.Join(root, name, stateJSON)
		stateData, _ := ioutil.ReadFile(stateFile)

		os.Stdout.Write(stateData)
		logrus.Debug(string(stateData))
		return nil
	},
}

func saveState(status string, container string, context *cli.Context) error {
	root := context.GlobalString("root")
	absRoot, _ := filepath.Abs(root)

	spec, err := setupSpec(context)
	if err != nil {
		fmt.Printf("setupSepc err\n")
		return err
	}

	rootfs, _ := filepath.Abs(spec.Root.Path)
	stateFile := filepath.Join(absRoot, container, stateJSON)
	cs := &specs.State{
		Version: spec.Version,
		ID:      context.Args().Get(0),
		Status:  status,
		Bundle:  rootfs,
	}
	stateData, _ := json.MarshalIndent(cs, "", "\t")

	if err := ioutil.WriteFile(stateFile, stateData, 0666); err != nil {
		panic(err)
	}

	return nil
}

func createContainer(container, bundle, stateRoot string, spec *specs.Spec) error {
	// Prepare container state directory
	stateDir := filepath.Join(stateRoot, container)
	_, err := os.Stat(stateDir)
	if err == nil {
		logrus.Errorf("Container %s exists", container)
		return fmt.Errorf("Container %s exists", container)
	}
	err = os.MkdirAll(stateDir, 0755)
	if err != nil {
		fmt.Printf("%s\n", err.Error())
		return err
	}
	defer func() {
		if err != nil {
			os.RemoveAll(stateDir)
		}
	}()

	return nil
}

func deleteContainer(root, container string) error {
	return os.RemoveAll(filepath.Join(root, container))
}
