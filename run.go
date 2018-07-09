package main

import (
	"fmt"
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
		if err := checkArgs(context, 1, exactArgs); err != nil {
			fmt.Printf("checkArgs err\n")
			return err
		}

		return cmdStartUnikernel(context)
	},
}

func printFile(path string, info os.FileInfo, err error) error {
    if err != nil {
        fmt.Print(err)
        return nil
    }
    fmt.Println(path)
    fmt.Printf("mode: %o\n", info.Mode)
    return nil
}


func cmdStartUnikernel(context *cli.Context) error {
	spec, err := setupSpec(context)
	if err != nil {
		fmt.Printf("setupSepc err\n")
		return err
	}

	rootfs,_ := filepath.Abs(spec.Root.Path)
//	fmt.Printf("exec %s at root %s cwd %s\n", spec.Process.Args, rootfs, spec.Process.Cwd)
//	fmt.Printf("context ==> %s\n", context)
//	fmt.Printf("spec ==> %s\n", spec)
//	filepath.Walk(rootfs, printFile)

	// call rexec
	saveState("running", context)
	os.Setenv("PATH", rootfs + ":" + rootfs + "/sbin:" + rootfs + "/bin:${PATH}")
	cmd := exec.Command("rexec", spec.Process.Args...)
	cmd.Dir = rootfs
	cmd.Env = append(os.Environ(),
		"RUMP_VERBOSE=1",
		"SUDO_UUID=1000",
	)


	stdout, err := cmd.StdoutPipe()
	if err != nil {
		panic(err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		panic(err)
	}

	if err := cmd.Start(); err != nil {
		panic(err)
	}


	go printOutputWithHeader(stdout, true)
	go printOutputWithHeader(stderr, true)

	if err := cmd.Wait(); err != nil {
		panic(err)
	}

	return err
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
