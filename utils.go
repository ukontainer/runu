package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	goruntime "runtime"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"

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
	runuAuxFileDir      = "/usr/lib/runu"
	fdInfoConfigJson    = "__RUMP_FDINFO_CONFIGJSON"
	fdInfoEnvPrefixNet  = "__RUMP_FDINFO_NET_"
	fdInfoEnvPrefixDisk = "__RUMP_FDINFO_DISK_"
	fdInfoEnvPrefixRoot = "__RUMP_FDINFO_ROOT"
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

func readPidFile(context *cli.Context, pidFile string) (int, error) {
	root := context.GlobalString("root")
	container := context.Args().Get(0)
	file := filepath.Join(root, container, pidFile)
	pid, err := ioutil.ReadFile(file)
	if err != nil {
		return 0, err
	}
	pidI, err := strconv.Atoi(string(pid))
	if err != nil {
		return 0, err
	}

	return pidI, nil

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

func isAlpineImage(rootfs string) bool {
	osRelease := rootfs + "/etc/os-release"

	f, err := os.Open(osRelease)
	if err != nil {
		return false
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		matched, _ := regexp.MatchString("Alpine Linux",
			scanner.Text())
		if matched {
			return true
		}
	}

	return false
}

func changeLdso(spec *specs.Spec, rootfs string) error {
	for _, env := range spec.Process.Env {
		if strings.HasPrefix(env, "RUNU_AUX_DIR=") {
			runuAuxFileDir = strings.TrimLeft(env, "RUNU_AUX_DIR=")
		}
	}

	// XXX: only for alpine
	// install frankenlibc-ed libc.so to the system one
	if goruntime.GOARCH == "amd64" {
		if err := copyFile(runuAuxFileDir+"/libc.so",
			rootfs+"/lib/ld-musl-x86_64.so.1", 0755); err != nil {
			return err
		}
	} else if goruntime.GOARCH == "arm" {
		if err := copyFile(runuAuxFileDir+"/libc.so",
			rootfs+"/lib/ld-musl-armhf.so.1", 0755); err != nil {
			return err
		}
	} else if goruntime.GOARCH == "arm64" {
		if err := copyFile(runuAuxFileDir+"/libc.so",
			rootfs+"/lib/ld-musl-aarch64.so.1", 0755); err != nil {
			return err
		}
	}

	// install frankenlibc-ed libc.so to the system one
	if err := copyFile(runuAuxFileDir+"/lkick",
		rootfs+"/bin/lkick", 0755); err != nil {
		return err
	}

	return nil
}

type lklInterface struct {
	MacAddr   string `json:"mac,omitempty"`
	V4Addr    string `json:"ip"`
	V4MaskLen string `json:"masklen"`
	V6Addr    string `json:"ipv6,omitempty"`
	V6MaskLen string `json:"masklen6,omitempty"`
	Name      string `json:"name"`
	Iftype    string `json:"type"`
	Offload   string `json:"offload,omitempty"`
}

type lklConfig struct {
	V4Gateway  string         `json:"gateway,omitempty"`
	Interfaces []lklInterface `json:"interfaces,omitempty"`
	Debug      string         `json:"debug,omitempty"`
	DelayMain  string         `json:"delay_main,omitempty"`
	SingleCpu  string         `json:"singlecpu,omitempty"`
	Sysctl     string         `json:"sysctl,omitempty"`
}

type lklIfInfo struct {
	ifAddrs []net.IPNet
	ifName  string
	v4Gw    net.IP
}

func generateLklJsonFile(lklJson string, lklJsonOut *string, spec *specs.Spec) (*lklIfInfo, error) {
	var config lklConfig

	// IPv4 address
	ifInfo, err := setupNetwork(spec)
	if err != nil {
		return nil, fmt.Errorf("failed to parse ipv4/ipv6 address: (%s)", err)
	}

	if ifInfo == nil {
		logrus.Warnf("no interface detected")
		*lklJsonOut = lklJson
		return nil, nil
	}

	logrus.Debugf(ifInfo.ifAddrs[0].String())

	v4masklen, _ := ifInfo.ifAddrs[0].Mask.Size()
	v4addr := ifInfo.ifAddrs[0].IP
	v4addr = v4addr.To4()
	v4gw := ifInfo.v4Gw.To4()

	// read user-specified file
	if lklJson != "" {
		bytes, err := ioutil.ReadFile(lklJson)
		if err != nil {
			logrus.Errorf("failed to read JSON file: %s (%s)",
				lklJson, err)
			panic(err)
		}

		// decode json
		if err := json.Unmarshal(bytes, &config); err != nil {
			logrus.Errorf("failed to decode JSON file: %s (%s)",
				lklJson, err)
			panic(err)
		}

		// only replace IPv4 address when the address is written
		// with "AUTO": otherwise use user-specified address
		if len(config.Interfaces) > 0 {
			if config.Interfaces[0].V4Addr == "AUTO" {
				config.Interfaces[0].V4Addr = v4addr.String()
				config.Interfaces[0].V4MaskLen = strconv.Itoa(v4masklen)
				config.V4Gateway = v4gw.String()
			}
		}
	} else {
		config = lklConfig{
			Debug: "1",
		}
		config.Interfaces = append(config.Interfaces,
			lklInterface{
				V4Addr:    v4addr.String(),
				V4MaskLen: strconv.Itoa(v4masklen),
				Name:      ifInfo.ifName,
				Iftype:    "rumpfd",
			})
		config.V4Gateway = v4gw.String()
	}

	outJson, err := json.MarshalIndent(config, "", "    ")
	if err != nil {
		logrus.Errorf("failed to encode JSON file: %s (%s)",
			lklJson, err)
		panic(err)
	}

	ioutil.WriteFile(*lklJsonOut, outJson, os.ModePerm)

	return ifInfo, nil
}

func parseEnvs(spec *specs.Spec, context *cli.Context, rootfs string) ([]string, map[*os.File]bool) {
	specEnv := []string{}
	fds := map[*os.File]bool{}
	fdNum := 3
	hasRootFs := false
	lklJson := ""
	// check if -v is specified
	_, use9pFs := os.LookupEnv("LKL_USE_9PFS")

	for _, env := range spec.Process.Env {
		// look for LKL_ROOTFS env for .img/.iso files
		if strings.HasPrefix(env, "LKL_ROOTFS=") {
			lklRootfs := strings.TrimLeft(env, "LKL_ROOTFS=")

			// if a file exists in local/host, copy and use it
			if _, err := os.Stat(lklRootfs); err == nil {
				copyFile(lklRootfs,
					rootfs+"/"+filepath.Base(lklRootfs),
					0644)
				lklRootfs = "/" + filepath.Base(lklRootfs)
			}

			fd, nonblock := openRootfsFd(rootfs + "/" + lklRootfs)
			fds[fd] = nonblock
			specEnv = append(specEnv, fdInfoEnvPrefixRoot+"="+strconv.Itoa(fdNum))
			fdNum++
			hasRootFs = true
			continue
		}
		// look for LKL_NET env for tap/macvtap devices
		if strings.HasPrefix(env, "LKL_NET=") {
			lklNet := strings.TrimLeft(env, "LKL_NET=")

			fd, nonblock := openNetFd(lklNet, spec.Process.Env)
			fds[fd] = nonblock
			specEnv = append(specEnv, fdInfoEnvPrefixNet+lklNet+"="+strconv.Itoa(fdNum))
			fdNum++
			continue
		}
		// look for LKL_CONFIG env for json file
		if strings.HasPrefix(env, "LKL_CONFIG=") {
			lklJson = strings.TrimLeft(env, "LKL_CONFIG=")
			copyFile(lklJson, rootfs+"/"+filepath.Base(lklJson), 0644)
			lklJson = rootfs + "/" + filepath.Base(lklJson)

			continue
		}

		// lookf for LKL_USE_9PFS
		if strings.HasPrefix(env, "LKL_USE_9PFS=") {
			use9pFs = true
		}

		// XXX: should exclude duplicated PATH variable in spec.Env since
		// it eliminates following values
		if !strings.HasPrefix(env, "PATH=") {
			specEnv = append(specEnv, env)
		}
	}

	// Set IPv4 addr/route from CNI info
	// and configure as lkl-$(container-id)-out.json file
	container := context.Args().First()
	clen := 10
	if len(container) < 10 {
		clen = len(container)
	}
	lklJsonOut := rootfs + "/" + "lkl-" + container[:clen] + "-out.json"

	ifInfo, err := generateLklJsonFile(lklJson, &lklJsonOut, spec)
	if err != nil {
		panic(err)
	}

	// XXX: eth0 should be somewhere else
	if ifInfo != nil {
		lklNet := "eth0"
		fd, nonblock := openNetFd(lklNet, spec.Process.Env)
		fds[fd] = nonblock
		specEnv = append(specEnv, fdInfoEnvPrefixNet+lklNet+"="+strconv.Itoa(fdNum))
		fdNum++
	}

	if lklJsonOut != "" {
		fd, nonblock := openJsonFd(lklJsonOut)
		fds[fd] = nonblock
		specEnv = append(specEnv, fdInfoConfigJson+"="+strconv.Itoa(fdNum))
		fdNum++
	}

	// start 9pfs server as a child process
	// if there is no rootfs disk image
	if !hasRootFs && use9pFs {
		childArgs := []string{"--9ps=" + rootfs + "/"}
		cmd := exec.Command(os.Args[0], childArgs[0:]...)
		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout
		if err := cmd.Start(); err != nil {
			panic(err)
		}
		// pid file for 9pfs server
		root := context.GlobalString("root")
		name := context.Args().Get(0)
		pidf := filepath.Join(root, name, pidFile9p)
		f, _ := os.OpenFile(pidf,
			os.O_RDWR|os.O_CREATE|os.O_EXCL|os.O_SYNC, 0666)
		_, _ = fmt.Fprintf(f, "%d", cmd.Process.Pid)
		f.Close()
		logrus.Debugf("Starting command %s, Env=%s, rootfs=%s\n",
			cmd.Args, cmd.Env, rootfs)

		time.Sleep(100 * time.Millisecond)
		fd, nonblock := connect9pfs()
		fds[fd] = nonblock
		specEnv = append(specEnv, "9PFS_FD="+strconv.Itoa(fdNum))
		specEnv = append(specEnv, "9PFS_MNT=/")
	}

	return specEnv, fds
}

func prepareUkontainer(context *cli.Context) (*exec.Cmd, error) {
	container := context.Args().Get(0)
	spec, err := setupSpec(context)
	if err != nil {
		logrus.Warnf("setupSepc err %s\n", err)
		return nil, err
	}

	rootfs, _ := filepath.Abs(spec.Root.Path)
	// open fds to pass to main programs later
	specEnv, fds := parseEnvs(spec, context, rootfs)

	// fixup ldso to a pulled image
	err = changeLdso(spec, rootfs)
	if err != nil {
		logrus.Warnf("ldso fixup error. skipping (%s)", err)
	}

	for _, node := range DefaultDevices {
		createDeviceNode(rootfs, node)
	}

	// call rexec
	os.Setenv("PATH", rootfs+":"+rootfs+
		"/sbin:"+rootfs+"/bin:/bin:/sbin:")

	cmd := exec.Command(spec.Process.Args[0], spec.Process.Args[1:]...)
	// XXX: need a better way to detect
	if isAlpineImage(rootfs) && goruntime.GOOS == "darwin" {
		logrus.Debugf("This is alpine linux image")
		cmd = exec.Command("lkick", spec.Process.Args[0:]...)
	}

	// do chroot(2) in rexec-ed processes
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: false,
	}
	cmd.Dir = "/"

	// on Linux, libc.so is replaced in a chrooted directory
	if goruntime.GOOS == "linux" {
		cmd.SysProcAttr.Chroot = rootfs

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
	}

	cmd.Env = append(specEnv, "PATH=/bin:/sbin:"+os.Getenv("PATH"))
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// XXX: need to sort cmd.Extrafiles to sync with frankenlibc
	fdkeys := make([]*os.File, 0)
	for k := range fds {
		fdkeys = append(fdkeys, k)
	}
	sort.SliceStable(fdkeys, func(i, j int) bool {
		return fdkeys[i].Fd() < fdkeys[j].Fd()
	})
	for k := range fdkeys {
		cmd.ExtraFiles = append(cmd.ExtraFiles, fdkeys[k])
	}

	cwd, _ := os.Getwd()
	logrus.Debugf("Starting command %s, Env=%s, cwd=%s, chroot=%s",
		cmd.Args, cmd.Env, cwd, rootfs)
	if err := cmd.Start(); err != nil {
		logrus.Errorf("cmd error %s (cmd=%s)", err, cmd.Args)
		panic(err)
	}

	// pid file for forked process from `runu create`
	// it'll be used later by `runu start`
	root := context.GlobalString("root")
	pidf := filepath.Join(root, container, pidFilePriv)
	f, _ := os.OpenFile(pidf,
		os.O_RDWR|os.O_CREATE|os.O_EXCL|os.O_SYNC, 0666)

	_, _ = fmt.Fprintf(f, "%d", cmd.Process.Pid)
	f.Close()

	logrus.Debugf("PID=%d to pid file %s",
		cmd.Process.Pid, pidf)

	proc, err := os.FindProcess(cmd.Process.Pid)
	if err != nil {
		return nil, fmt.Errorf("couldn't find pid %d", cmd.Process.Pid)
	}
	proc.Signal(syscall.Signal(syscall.SIGSTOP))

	// XXX:
	// os/exec.Start() close and open extrafiles thus strip O_NONBLOCK flag
	// thus re-enable it here
	for fd, nbFlag := range fds {
		if err := syscall.SetNonblock(int(fd.Fd()), nbFlag); err != nil {
			logrus.Errorf("setNonBlock %d error: %s\n", int(fd.Fd()), err)
			panic(err)
		}
	}

	return cmd, nil
}
