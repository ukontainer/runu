module github.com/ukontainer/runu

go 1.13

replace github.com/Sirupsen/logrus v1.4.2 => github.com/sirupsen/logrus v1.4.2

replace github.com/containerd/containerd v1.4.0 => github.com/ukontainer/containerd v1.2.1-0.20201214074721-d67c6abf4ce2

require (
	github.com/containerd/console v1.0.1
	github.com/containerd/containerd v1.4.0
	github.com/containerd/go-runc v0.0.0-20200930051345-7c5957f67ef4
	github.com/containerd/typeurl v1.0.1
	github.com/docker/go-p9p v0.0.0-20191112112554-37d97cf40d03
	github.com/gogo/protobuf v1.3.1
	github.com/hashicorp/go-multierror v1.1.0 // indirect
	github.com/opencontainers/runtime-spec v1.0.3-0.20200728170252-4d89ac9fbff6
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.7.0
	github.com/urfave/cli v1.22.2
	github.com/vishvananda/netlink v1.1.0
	github.com/vishvananda/netns v0.0.0-20200728191858-db3c7e526aae
	golang.org/x/sys v0.0.0-20201202213521-69691e467435
	google.golang.org/grpc v1.31.1 // indirect
)
