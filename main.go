package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/opencontainers/runtime-spec/specs-go"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

const (
	specConfig = "config.json"
	usage = "runu run [ -b bundle ] <container-id>"
)

func main() {
	app := cli.NewApp()
	app.Name = "runu"
	app.Usage = usage

	var v []string
	v = append(v, fmt.Sprintf("spec: %s", specs.Version))
	app.Version = strings.Join(v, "\n")

	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "debug",
			Usage: "enable debug output for logging",
		},
		cli.StringFlag{
			Name:  "log",
			Value: "/dev/null",
			Usage: "set the log file path where internal debug information is written",
		},
		cli.StringFlag{
			Name:  "log-format",
			Value: "text",
			Usage: "set the format used by logs ('text' (default), or 'json')",
		},
	}
	app.Commands = []cli.Command{
		runCommand,
	}
	app.Before = func(context *cli.Context) error {
		if context.GlobalBool("debug") {
			logrus.SetLevel(logrus.DebugLevel)
		}
		if path := context.GlobalString("log"); path != "" {
			f, err := os.OpenFile(path,
				os.O_CREATE|os.O_WRONLY|os.O_APPEND|os.O_SYNC,
				0666)
			if err != nil {
				return err
			}
			logrus.SetOutput(f)
		}
		switch context.GlobalString("log-format") {
		case "text":
			// retain logrus's default.
		case "json":
			logrus.SetFormatter(new(logrus.JSONFormatter))
		default:
			return fmt.Errorf("unknown log-format %q",
				context.GlobalString("log-format"))
		}
		return nil
	}

	if err := app.Run(os.Args); err != nil {
		panic(err)
	}
}

