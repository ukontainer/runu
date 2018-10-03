package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"bufio"
	"io"
	"encoding/json"
	_ "time"

	"github.com/opencontainers/runc/libcontainer"
	"github.com/opencontainers/runtime-spec/specs-go"
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


func printOutputWithHeader(r io.Reader, verbose bool) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		if verbose {
			fmt.Printf("%s\n", scanner.Text())
			logrus.Info("%s\n", scanner.Text())
		}
	}
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
			fmt.Printf("error: dir not found (%s)\n", bundle)
			return nil, err
		}
	}
	spec, err := loadSpec(specConfig)
	if err != nil {
		fmt.Printf("loadSpec err (%s)\n", err)
		return nil, err
	}
	return spec, nil
}

// newProcess returns a new libcontainer Process with the arguments from the
// spec and stdio from the current process.
func newProcess(p specs.Process) (*libcontainer.Process, error) {
	lp := &libcontainer.Process{
		Args: append([]string{"rexec"}, p.Args[0:]...),
		Env:  p.Env,
		// TODO: fix libcontainer's API to better support uid/gid in a typesafe way.
		User:            fmt.Sprintf("%d:%d", p.User.UID, p.User.GID),
		Cwd:             p.Cwd,
		Label:           p.SelinuxLabel,
		NoNewPrivileges: &p.NoNewPrivileges,
		AppArmorProfile: p.ApparmorProfile,
	}

	if p.ConsoleSize != nil {
		lp.ConsoleWidth = uint16(p.ConsoleSize.Width)
		lp.ConsoleHeight = uint16(p.ConsoleSize.Height)
	}

	var procAttr os.ProcAttr
	procAttr.Files = []*os.File{os.Stdin,
	    		    os.Stdout, os.Stderr}
	procAttr.Env = os.Environ()
	binpath, err := exec.LookPath("rexec")
	if err != nil {
	   logrus.Printf("lookpath err=%s\n", err)
	}
	lp2, err := os.StartProcess(binpath, p.Args, &procAttr)
	if err != nil {
	   logrus.Printf("StartProcess err=%s %s\n", err, lp2)
	}
	logrus.Printf("StartProcess %s pid=%d\n", lp2, lp2.Pid)

	pid, err := lp.Pid()
	if err != nil {
	   logrus.Printf("getpid err=%s\n", err)
	}

	logrus.Printf("new proc pid=%d\n", pid)
	return lp, nil
}

func startUnikernel(context *cli.Context) error {
/*
	fd, err := os.OpenFile(pipeFile, os.O_RDWR, os.ModeNamedPipe)
	if err != nil {
     		logrus.Printf("Open named pipe file error:", err)
		panic(err)
	}

	reader := bufio.NewReader(fd)
	logrus.Printf("wait to read named pipe file %s\n", fd)
	_, err = reader.ReadByte()
	logrus.Printf("done read named pipe file\n")
	if err == nil {
		logrus.Print("go rexec\n")
	} else {
		panic(err)
	}
*/

	spec, err := setupSpec(context)
	if err != nil {
		logrus.Printf("setupSepc err\n")
		return err
	}

	rootfs,_ := filepath.Abs(spec.Root.Path)
//	logrus.Printf("exec %s at root %s cwd %s\n", spec.Process.Args, rootfs, spec.Process.Cwd)
//	logrus.Printf("context ==> %s\n", context)
//	logrus.Printf("spec ==> %s\n", spec)

	// call rexec
	os.Setenv("PATH", rootfs + ":" + rootfs + "/sbin:" + rootfs + "/bin:${PATH}")

/*
	logrus.Printf("PATH=%s\n", os.Getenv("PATH"))
	newProcess(*spec.Process)
	return nil
*/
	cmd := exec.Command("rexec", spec.Process.Args...)
	cmd.Dir = rootfs
	cmd.Env = append(os.Environ(),
		"RUMP_VERBOSE=1",
		"PYTHONHOME=/python",
		"HOME=/",
		"SUDO_UUID=1000",
		"LKL_OFFLOAD=1",
		"LKL_BOOT_CMDLINE=mem=1G",
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
		fmt.Printf("cmd error %s\n", err)
		panic(err)
	}

	// write pid file
	if pidf := context.String("pid-file"); pidf != "" {
		f, err := os.OpenFile(pidf, os.O_RDWR|os.O_CREATE|os.O_EXCL|os.O_SYNC, 0666)
		if err != nil {
			fmt.Printf("ERR: %s\n", err)
			return err
		}
		_, err = fmt.Fprintf(f, "%d", cmd.Process.Pid)
		f.Close()
	}

	saveState("running", cmd.Process.Pid, context)

	go printOutputWithHeader(stdout, true)
	go printOutputWithHeader(stderr, true)

	if err := cmd.Wait(); err != nil {
		fmt.Printf("%s\n", err)
		panic(err)
	}

	saveState("stopped", cmd.Process.Pid, context)
	return nil
}
