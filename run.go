package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/urfave/cli"
)

// default action is to start a container
var runCommand = cli.Command{
	Name:  "run",
	Usage: "create and run a container",
	ArgsUsage: `<container-id>

Where "<container-id>" is your name for the instance of the container that you
are starting. The name you provide for the container instance must be unique on
your host.`,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "bundle, b",
			Value: "",
			Usage: `path to the root of the bundle directory, defaults to the current directory`,
		},
	},
	Action: func(context *cli.Context) error {
		if err := checkArgs(context, 1, exactArgs); err != nil {
			return err
		}
		spec, err := setupSpec(context)
		if err != nil {
			return err
		}

		fmt.Printf("exec %s\n", spec.Process.Args)

		// call rexec
		cmd := exec.Command("rexec", spec.Process.Args[0:]...)
		cmd.Env = append(os.Environ(),
			"RUMP_VERBOSE=1",
		)

		stdout, err := cmd.StdoutPipe()
		if err != nil {
			log.Fatal(err)
		}
		if err := cmd.Start(); err != nil {
			log.Fatal(err)
		}

		go printOutputWithHeader(stdout, true)

		if err := cmd.Wait(); err != nil {
			log.Fatal(err)
		}

//		sout, err := cmd.CombinedOutput()
//		fmt.Println(string(sout))
//		fmt.Println(err)

		return err
	},
}
