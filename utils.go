package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
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

func prepareUkontainer(context *cli.Context) error {
	name := context.Args().Get(0)
	spec, err := setupSpec(context)
	if err != nil {
		logrus.Warn("setupSepc err\n")
		return err
	}

	rootfs, _ := filepath.Abs(spec.Root.Path)
	// call rexec
	os.Setenv("PATH", rootfs+":"+rootfs+
		"/sbin:"+rootfs+"/bin")

	cmd := exec.Command("rexec", spec.Process.Args...)
	cmd.Dir = rootfs

	// XXX: should exclude duplicated PATH variable in spec.Env since
	// it eliminates following values
	specEnv := []string{}
	for _, env := range spec.Process.Env {
		if !strings.HasPrefix(env, "PATH=") {
			specEnv = append(specEnv, env)
		}
	}
	cmd.Env = append(os.Environ(), specEnv...)
	logrus.Debugf("Starting command %s, PATH= %s\n", cmd.Args, cmd.Env)

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		fmt.Printf("cmd error %s\n", err)
		panic(err)
	}

	// write pid file
	if pidf := context.String("pid-file"); pidf != "" {
		// 0) pid file for containerd
		f, err := os.OpenFile(pidf,
			os.O_RDWR|os.O_CREATE|os.O_EXCL|os.O_SYNC, 0666)
		if err != nil {
			fmt.Printf("ERR: %s\n", err)
			return err
		}
		_, _ = fmt.Fprintf(f, "%d", cmd.Process.Pid)
		f.Close()

	}
	// 1) pid file for runu itself
	root := context.GlobalString("root")
	pidf := filepath.Join(root, name, pidFilePriv)
	f, _ := os.OpenFile(pidf,
		os.O_RDWR|os.O_CREATE|os.O_EXCL|os.O_SYNC, 0666)

	_, _ = fmt.Fprintf(f, "%d", cmd.Process.Pid)
	f.Close()

	logrus.Debugf("PID=%d to pid file %s",
		cmd.Process.Pid, pidf)

	proc, err := os.FindProcess(cmd.Process.Pid)
	if err != nil {
		return fmt.Errorf("couldn't find pid %d", cmd.Process.Pid)
	}
	proc.Signal(syscall.Signal(syscall.SIGSTOP))

	return nil
}
