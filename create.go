package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

var createCommand = cli.Command{
	Name:      "create",
	Usage:     "create a container",
	ArgsUsage: `<container-id>`,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "bundle, b",
			Value: "",
			Usage: `path to the root of the bundle directory, defaults to the current directory`,
		},
		cli.StringFlag{
			Name:  "pid-file",
			Usage: "specify the file to write the process id to",
		},
	},
	Action: func(context *cli.Context) error {
		args := context.Args()
		if args.Present() == false {
			return fmt.Errorf("Missing container ID")
		}

		return cmdCreateUkon(context, false)
	},
}

func cmdCreateUkon(context *cli.Context, attach bool) error {
	root := context.GlobalString("root")
	bundle := context.String("bundle")
	container := context.Args().First()
	ocffile := filepath.Join(bundle, specConfig)
	spec, err := loadSpec(ocffile)
	const stdioFdCount = 3

	if err != nil {
		return fmt.Errorf("load config failed: %v", err)
	}
	if container == "" {
		return fmt.Errorf("no container id provided")
	}

	err = createContainer(container, bundle, root, spec)
	if err != nil {
		return fmt.Errorf("failed to create container: %v", err)
	}

	// call `runu boot` to create new process
	self, err := os.Executable()
	if err != nil {
		return fmt.Errorf("could not identify who am I: %v", err)
	}

	args := []string{}
	if val := context.GlobalString("log-format"); val != "" {
		args = append(args, "-log-format")
		args = append(args, context.GlobalString("log-format"))
	}
	if val := context.GlobalString("log"); val != "" {
		args = append(args, "-log")
		args = append(args, context.GlobalString("log"))
	}
	if val := context.GlobalString("root"); val != "" {
		args = append(args, "-root")
		args = append(args, context.GlobalString("root"))
	}
	if context.GlobalBool("debug") {
		args = append(args, "-debug")
	}
	args = append(args, "boot")
	if val := context.String("bundle"); val != "" {
		args = append(args, "-bundle")
		args = append(args, context.String("bundle"))
	}
	if val := context.String("pid-file"); val != "" {
		args = append(args, "-pid-file")
		args = append(args, context.String("pid-file"))
	}
	args = append(args, container)

	cmd := exec.Command(self, args...)

	// wait for init complete
	parentPipe, childPipe, err := os.Pipe()
	if err != nil {
		return fmt.Errorf("create: pipe failure (%s)", err)
	}
	cmd.ExtraFiles = append(cmd.ExtraFiles, childPipe)
	cmd.Env = append(cmd.Env,
		fmt.Sprintf("_LIBCONTAINER_INITPIPE=%d", stdioFdCount+len(cmd.ExtraFiles)-1),
	)

	cwd, _ := os.Getwd()
	logrus.Debugf("Starting command %s, cwd=%s, root=%s",
		cmd.Args, cwd, root)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: false,
	}

	if err := cmd.Start(); err != nil {
		panic(err)
	}

	go func() {
		err = cmd.Wait()
		if err != nil {
			fmt.Printf("failed to wait a process: %s", cmd.Args)
			panic(err)
		}
	}()

	buf := make([]byte, 1)
	logrus.Debugf("Waiting for pipe to complete boot %s", cmd.Args)
	if _, err := parentPipe.Read(buf); err != nil {
		fmt.Printf("pipe read: %s", err)
	}
	parentPipe.Close()
	childPipe.Close()
	logrus.Debugf("Waiting pipe done")

	return nil
}
