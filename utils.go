package main

import (
	"fmt"
	"os"
	"bufio"
	"io"

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
