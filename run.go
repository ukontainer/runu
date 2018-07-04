package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"encoding/json"

	"github.com/urfave/cli"
	"github.com/opencontainers/runtime-spec/specs-go"
)

// default action is to start a container
var runCommand = cli.Command{
	Name:  "run",
	Usage: "create and run a container",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "bundle, b",
			Value: "",
			Usage: `path to the root of the bundle directory, defaults to the current directory`,
		},
	},
	Action: func(context *cli.Context) error {
		if err := checkArgs(context, 2, exactArgs); err != nil {
			fmt.Printf("checkArgs err\n")
			return err
		}
		spec, err := setupSpec(context)
		if err != nil {
			fmt.Printf("setupSepc err\n")
			return err
		}

		fmt.Printf("exec %s\n", spec.Process.Args)
		rootfs,_ := filepath.Abs(spec.Root.Path)

		// call rexec
		os.Setenv("PATH", "${PATH}:" + rootfs + "/sbin:" + rootfs + "/bin")
		cmd := exec.Command("rexec", context.Args()...)
		cmd.Dir = spec.Process.Cwd
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

		return err
	},
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
			return nil, err
		}
	}
	spec, err := loadSpec(specConfig)
	if err != nil {
		return nil, err
	}
	return spec, nil
}
