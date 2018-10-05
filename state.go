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

var stateCommand = cli.Command{
	Name:  "state",
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
		logrus.Print(string(stateData))
		return nil
	},
}

func saveState(status string, name string, context *cli.Context) error {
	root := context.GlobalString("root")
	absRoot, err := filepath.Abs(root)

	spec, err := setupSpec(context)
	if err != nil {
		fmt.Printf("setupSepc err\n")
		return err
	}

	rootfs,_ := filepath.Abs(spec.Root.Path)
	stateFile := filepath.Join(absRoot, name, stateJSON)
	cs := &specs.State {
		Version: spec.Version,
		ID: context.Args().Get(0),
		Status: status,
		Bundle: rootfs,
	}
	stateData, _ := json.MarshalIndent(cs, "", "\t")

	if err := ioutil.WriteFile(stateFile, stateData, 0666); err != nil {
		panic(err);
		return err
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
	err = os.MkdirAll(stateDir, 0644)
	if err != nil {
		logrus.Infof("%s", err.Error())
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
