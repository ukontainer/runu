package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/urfave/cli"
	"github.com/sirupsen/logrus"
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
//	logrus.Printf("exec %s at root %s cwd %s\n", spec.Process.Args, rootfs, spec.Process.Cwd)
//	logrus.Printf("context ==> %s\n", context)
//	logrus.Printf("spec ==> %s\n", spec)

	// call rexec
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

	saveState("running", cmd.Process.Pid, context)

	go printOutputWithHeader(stdout, true)
	go printOutputWithHeader(stderr, true)

	if err := cmd.Wait(); err != nil {
		logrus.Printf("%s\n", err)
		panic(err)
	}

	saveState("stopped", cmd.Process.Pid, context)
	return nil
}
