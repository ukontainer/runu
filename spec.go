package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/opencontainers/runtime-spec/specs-go"
	"github.com/urfave/cli"
)

// Example generates a template for spec file
// inherited from runc/libcontainer/specconv/example.go to
// avoid cgroup dependency
func Example() *specs.Spec {
	spec := &specs.Spec{
		Version: specs.Version,
		Root: &specs.Root{
			Path:     "rootfs",
			Readonly: true,
		},
		Process: &specs.Process{
			Terminal: true,
			User:     specs.User{},
			Args: []string{
				"sh",
			},
			Env: []string{
				"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin",
				"TERM=xterm",
			},
			Cwd:             "/",
			NoNewPrivileges: true,
		},
		Hostname: "runu",
		Linux: &specs.Linux{
			MaskedPaths: []string{
				"/proc/acpi",
				"/proc/asound",
				"/proc/kcore",
				"/proc/keys",
				"/proc/latency_stats",
				"/proc/timer_list",
				"/proc/timer_stats",
				"/proc/sched_debug",
				"/sys/firmware",
				"/proc/scsi",
			},
			ReadonlyPaths: []string{
				"/proc/bus",
				"/proc/fs",
				"/proc/irq",
				"/proc/sys",
				"/proc/sysrq-trigger",
			},
			Resources: &specs.LinuxResources{
				Devices: []specs.LinuxDeviceCgroup{
					{
						Allow:  false,
						Access: "rwm",
					},
				},
			},
			Namespaces: []specs.LinuxNamespace{
				{
					Type: specs.PIDNamespace,
				},
				{
					Type: specs.NetworkNamespace,
				},
				{
					Type: specs.IPCNamespace,
				},
				{
					Type: specs.UTSNamespace,
				},
				{
					Type: specs.MountNamespace,
				},
			},
		},
	}
	return spec
}

var specCommand = cli.Command{
	Name:        "spec",
	Usage:       "create a new specification file",
	ArgsUsage:   "",
	Description: `The spec command creates the new specification file named "` + specConfig + `" for the bundle.`,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "bundle, b",
			Value: "",
			Usage: "path to the root of the bundle directory",
		},
	},
	Action: func(context *cli.Context) error {
		if err := checkArgs(context, 0, exactArgs); err != nil {
			return err
		}
		spec := Example()

		checkNoFile := func(name string) error {
			_, err := os.Stat(name)
			if err == nil {
				return fmt.Errorf("File %s exists. Remove it first", name)
			}
			if !os.IsNotExist(err) {
				return err
			}
			return nil
		}
		bundle := context.String("bundle")
		if bundle != "" {
			if err := os.Chdir(bundle); err != nil {
				return err
			}
		}
		if err := checkNoFile(specConfig); err != nil {
			return err
		}
		data, err := json.MarshalIndent(spec, "", "\t")
		if err != nil {
			return err
		}

		return ioutil.WriteFile(specConfig, data, 0666)
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
			return nil, fmt.Errorf("spec: dir not found (bundle=%s,%v)", bundle, err)
		}
	}
	spec, err := loadSpec(specConfig)
	if err != nil {
		fmt.Printf("loadSpec err (%s)\n", err)
		return nil, err
	}
	return spec, nil
}
