package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	goruntime "runtime"
	"strconv"
	"strings"
	"syscall"

	"github.com/opencontainers/runtime-spec/specs-go"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

const (
	exactArgs = iota
	minArgs
	maxArgs
)

var (
	runuAuxFileDir = "/usr/local/lib/runu"
	FDINFO_NAME_CONFIGJSON =  "__RUMP_FDINFO_CONFIGJSON"
	FDINFO_ENV_PREFIX_NET = "__RUMP_FDINFO_NET_"
	FDINFO_ENV_PREFIX_DISK =  "__RUMP_FDINFO_DISK_"
	FDINFO_ENV_PREFIX_ROOT = "__RUMP_FDINFO_ROOT"
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

func copyFile(src, dst string, mode os.FileMode) error {
	b, err := ioutil.ReadFile(src)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(dst, b, mode)
	if err != nil {
		return err
	}
	return nil
}

func changeLdso(spec *specs.Spec, rootfs string) error {
	for _, env := range spec.Process.Env {
		if strings.HasPrefix(env, "RUNU_AUX_DIR=") {
			runuAuxFileDir = strings.TrimLeft(env, "RUNU_AUX_DIR=")
		}
	}

	// XXX: only for alpine
	// install frankenlibc-ed libc.so to the system one
	if err := copyFile(runuAuxFileDir+"/libc.so",
		rootfs+"/lib/ld-musl-x86_64.so.1", 0755); err != nil {
		return err
	}

	return nil
}

func parseEnvs(spec *specs.Spec, rootfs string) ([]string, []*os.File) {
	specEnv := []string{}
	fds := []*os.File{}
	fdNum := 3

	for _, env := range spec.Process.Env {
		// look for LKL_ROOTFS env for .img/.iso files
		if strings.HasPrefix(env, "LKL_ROOTFS=") {
			lklRootfs := strings.TrimLeft(env, "LKL_ROOTFS=")
			fd := openRootfsFd(rootfs+"/"+lklRootfs)
			fds = append(fds, fd)
			specEnv = append(specEnv, FDINFO_ENV_PREFIX_ROOT+"=" + strconv.Itoa(fdNum))
			fdNum++
			continue
		}
		// look for LKL_NET env for tap/macvtap devices
		if strings.HasPrefix(env, "LKL_NET=") {
			lklNet := strings.TrimLeft(env, "LKL_NET=")

			fd := openNetFd(lklNet, spec.Process.Env)
			fds = append(fds, fd)
			specEnv = append(specEnv, FDINFO_ENV_PREFIX_NET+lklNet+"=" + strconv.Itoa(fdNum))
			fdNum++
			continue
		}
		// look for LKL_CONFIG env for json file
		if strings.HasPrefix(env, "LKL_CONFIG=") {
			lklJson := strings.TrimLeft(env, "LKL_CONFIG=")
			copyFile(lklJson, rootfs+"/"+filepath.Base(lklJson), 0644)
			lklJson = "/" + filepath.Base(lklJson)

			fd := openJsonFd(rootfs+"/"+lklJson)
			fds = append(fds, fd)
			specEnv = append(specEnv, FDINFO_NAME_CONFIGJSON +"=" + strconv.Itoa(fdNum))
			fdNum++
			continue
		}

		// XXX: should exclude duplicated PATH variable in spec.Env since
		// it eliminates following values
		if !strings.HasPrefix(env, "PATH=") {
			specEnv = append(specEnv, env)
		}
	}


	return specEnv, fds
}

func prepareUkontainer(context *cli.Context) error {
	name := context.Args().Get(0)
	spec, err := setupSpec(context)
	if err != nil {
		logrus.Warn("setupSepc err\n")
		return err
	}

	rootfs, _ := filepath.Abs(spec.Root.Path)
	// open fds to pass to main programs later
	specEnv,fds := parseEnvs(spec, rootfs)

	// fixup ldso to a pulled image
	err = changeLdso(spec, rootfs)
	if err != nil {
		logrus.Warnf("ldso fixup error. skipping (%s)", err)
	}

	for _, node := range DefaultDevices {
		createDeviceNode(rootfs, node)
	}

	// call rexec
	os.Setenv("PATH", "/bin:/sbin:"+rootfs+":"+rootfs+
		"/sbin:"+rootfs+"/bin")

	cmd := exec.Command(spec.Process.Args[0], spec.Process.Args[1:]...)

	if goruntime.GOOS == "linux" {
		// do chroot(2) in rexec-ed processes
		cmd.SysProcAttr = &syscall.SysProcAttr{
			Chroot: rootfs,
		}
		cmd.Dir = "/"

		binpath, err := exec.LookPath(spec.Process.Args[0])
		if err != nil {
			logrus.Errorf("cmd %s not found %s",
				spec.Process.Args[0], err)
			os.Setenv("PATH", "/bin:/sbin:/usr/bin:"+rootfs+":"+rootfs+
				"/sbin:"+rootfs+"/bin")
		}

		if strings.HasPrefix(binpath, rootfs) {
			cmd.Path = strings.Split(binpath, rootfs)[1]
			logrus.Debugf("cmd %s found at %s=>%s",
				spec.Process.Args[0], binpath, cmd.Path)
		}

	} else if goruntime.GOOS == "darwin" {
		// XXX: because it's *hard* to create static-linked
		// binary (rexec), we don't use chroot on OSX.  maybe
		// eliminating rexec and do the job by runu should be better
		cmd.Dir = rootfs
	}
	cmd.Env = append(specEnv, "PATH=/bin:/sbin:"+os.Getenv("PATH"))
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.ExtraFiles = fds

	cwd, _ := os.Getwd()
	logrus.Debugf("Starting command %s, Env=%s, cwd=%s, chroot=%s",
		spec.Process.Args, cmd.Env, cwd, rootfs)
	if err := cmd.Start(); err != nil {
		logrus.Errorf("cmd error %s (cmd=%s)", err, cmd)
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

	// XXX:
	// os/exec.Start() close and open extrafiles thus strip O_NONBLOCK flag
	// thus re-enable it here
	for _, fd := range fds {
		if err := syscall.SetNonblock(int(fd.Fd()), true); err != nil {
			logrus.Errorf("setNonBlock %d error: %s\n", int(fd.Fd()), err)
			panic(err)
		}
	}

	return nil
}
