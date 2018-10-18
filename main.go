package main

import (
	"fmt"
	"os"
	"strings"

	goruntime "runtime"
	"github.com/opencontainers/runtime-spec/specs-go"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

const (
	specConfig = "config.json"
	stateJSON = "state.json"
	usage = "runu run [ -b bundle ] <container-id>"
	arch = goruntime.GOARCH
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
			Value: "/tmp/runu.log",
			Usage: "set the log file path where internal debug information is written",
		},
		cli.StringFlag{
			Name:  "log-format",
			Value: "text",
			Usage: "set the format used by logs ('text' (default), or 'json')",
		},
		cli.StringFlag{
			Name:  "root",
			Value: "/run/runu",
			Usage: "root directory for storage of container state (this should be located in tmpfs)",
		},
	}
	app.Commands = []cli.Command{
		createCommand,
		runCommand,
		// XXX: unimplemented
		startCommand,
		stateCommand,
		execCommand,
		killCommand,
		deleteCommand,
	}

	app.Before = func(context *cli.Context) error {
		if path := context.GlobalString("log"); path != "" {
			f, err := os.OpenFile(path,
				os.O_CREATE|os.O_WRONLY|os.O_APPEND|os.O_SYNC,
				0666)
			if err != nil {
				fmt.Printf("%s\n", err)
				return err
			}
			logrus.SetOutput(f)
		}
		if context.GlobalBool("debug") {
			logrus.SetLevel(logrus.DebugLevel)
			logrus.SetOutput(os.Stdout)
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

		err := handleSystemLog("","")
		if err != nil {
			return err
		}
		logrus.Printf("Runu called with args: %v\n", os.Args)
		return nil
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Printf("%s\n", err)
	}

}

