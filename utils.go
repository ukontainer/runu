package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

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


func prepareUkontainer(context *cli.Context) error {
	name := context.Args().Get(0)
	spec, err := setupSpec(context)
	if err != nil {
		logrus.Printf("setupSepc err\n")
		return err
	}

	rootfs,_ := filepath.Abs(spec.Root.Path)
	// call rexec
	os.Setenv("PATH", rootfs + ":" + rootfs +
		"/sbin:" + rootfs + "/bin:${PATH}")

	cmd := exec.Command("rexec", spec.Process.Args...)
	cmd.Dir = rootfs
	// XXX: should exclude Env[0](PATH=..) since
	// it eliminates following values
	cmd.Env = append(os.Environ(), spec.Process.Env[1:]...)

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
		_, err = fmt.Fprintf(f, "%d", cmd.Process.Pid)
		f.Close()

	}
	// 1) pid file for runu itself
	root := context.GlobalString("root")
	pidf := filepath.Join(root, name, "runu.pid")
	f, err := os.OpenFile(pidf,
		os.O_RDWR|os.O_CREATE|os.O_EXCL|os.O_SYNC, 0666)

	_, err = fmt.Fprintf(f, "%d", cmd.Process.Pid)
	f.Close()

	logrus.Printf("PID=%d to pid file %s",
		cmd.Process.Pid, pidf)

	proc, _ := os.FindProcess(cmd.Process.Pid)
	proc.Signal(syscall.Signal(syscall.SIGSTOP))

	saveState("running", name, context)

	go func() {
		if err := cmd.Wait(); err != nil {
			waitstatus := cmd.ProcessState.Sys().(syscall.WaitStatus)
			fmt.Printf("%s\n", err)
			if !waitstatus.Signaled() {
				panic(err)
			}
		}

		saveState("stopped", name, context)
	}()

	return nil
}
