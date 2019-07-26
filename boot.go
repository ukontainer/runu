package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	goruntime "runtime"
	"strconv"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"golang.org/x/sys/unix"
)

var bootCommand = cli.Command{
	Name:  "boot",
	Usage: "(internal use only) boot a container",
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
		return bootContainer(context, false)
	},
}

func bootContainer(context *cli.Context, attach bool) error {
	container := context.Args().First()
	cmd, err := prepareUkontainer(context)
	if err != nil {
		return fmt.Errorf("failed to prepare container: %v", err)
	}

	// write pid file for containerd-shim
	if pidf := context.String("pid-file"); pidf != "" {
		// 0) pid file for containerd
		f, err := os.OpenFile(pidf,
			os.O_RDWR|os.O_CREATE|os.O_EXCL|os.O_SYNC, 0666)
		if err != nil {
			return fmt.Errorf("pid-file: %s\n", err)
		}
		//
		// XXX:
		// linux should tell containerd with a child process (subreaper)
		// while darwin should with a grandchild process (ReapMore)
		//
		if goruntime.GOOS == "linux" {
			_, _ = fmt.Fprintf(f, "%d", os.Getpid())
		} else if goruntime.GOOS == "darwin" {
			_, _ = fmt.Fprintf(f, "%d", cmd.Process.Pid)
		}
		f.Close()
	}

	saveState("created", container, context)

	envInitPipe := os.Getenv("_LIBCONTAINER_INITPIPE")
	pipefd, err := strconv.Atoi(envInitPipe)
	if err != nil {
		return fmt.Errorf("unable to convert _LIBCONTAINER_INITPIPE=%s to int: %s", envInitPipe, err)
	}

	// notify to `runu create`
	pipe := os.NewFile(uintptr(pipefd), "pipe")
	logrus.Debugf("writing pipe %s _LIBCONTAINER_INITPIPE=%s", pipe.Name(), envInitPipe)
	pipe.Write([]byte("1"))
	pipe.Close()

	// catch child errors if possible
	err = cmd.Wait()

	// signal to containerd-shim to request exit
	bundle := context.String("bundle")
	file := filepath.Join(bundle, "shim.pid")
	pid, _ := ioutil.ReadFile(file)
	pidI, _ := strconv.Atoi(string(pid))
	unix.Kill(pidI, unix.SIGCHLD)
	logrus.Debugf("sending SIGCHLD to parent %d", pidI)

	saveState("stopped", container, context)
	logrus.Debugf("process stopped %s", cmd.Args)

	return err
}
