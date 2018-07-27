package main

import (
	"fmt"
	"os"
	"bufio"
	"io"
	"encoding/json"

	"github.com/opencontainers/runtime-spec/specs-go"
	"github.com/urfave/cli"
	"github.com/sirupsen/logrus"
)

const (
	exactArgs = iota
	minArgs
	maxArgs
)

func checkArgs(context *cli.Context, expected, checkType int) error {
	var err error
	cmdName := context.Command.Name
	switch checkType {
	case exactArgs:
		if context.NArg() != expected {
			err = fmt.Errorf(
				"%s: %q requires exactly %d argument(s)",
				os.Args[0], cmdName, expected)
		}
	case minArgs:
		if context.NArg() < expected {
			err = fmt.Errorf(
				"%s: %q requires a minimum of %d argument(s)",
				os.Args[0], cmdName, expected)
		}
	case maxArgs:
		if context.NArg() > expected {
			err = fmt.Errorf(
				"%s: %q requires a maximum of %d argument(s)",
				os.Args[0], cmdName, expected)
		}
	}

	if err != nil {
		fmt.Printf("Incorrect Usage.\n\n")
		cli.ShowCommandHelp(context, cmdName)
		return err
	}
	return nil
}


func printOutputWithHeader(r io.Reader, verbose bool) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		if verbose {
			fmt.Printf("%s\n", scanner.Text())
			logrus.Info("%s\n", scanner.Text())
		}
	}
}

// loadSpec loads the specification from the provided path.
func loadSpec(cPath string) (spec *specs.Spec, err error) {
	cf, err := os.Open(cPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("JSON specification file %s not found", cPath)
		}
		return nil, err
	}
	defer cf.Close()

	if err = json.NewDecoder(cf).Decode(&spec); err != nil {
		return nil, err
	}
	return spec, nil
}

// setupSpec performs initial setup based on the cli.Context for the container
func setupSpec(context *cli.Context) (*specs.Spec, error) {
	bundle := context.String("bundle")
	if bundle != "" {
		if err := os.Chdir(bundle); err != nil {
			fmt.Printf("error: dir not found (%s)\n", bundle)
			return nil, err
		}
	}
	spec, err := loadSpec(specConfig)
	if err != nil {
		fmt.Printf("loadSpec err (%s)\n", err)
		return nil, err
	}
	return spec, nil
}
