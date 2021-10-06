module github.com/ukontainer/runu

go 1.13

replace github.com/Sirupsen/logrus v1.4.2 => github.com/sirupsen/logrus v1.4.2

replace (
	github.com/containerd/containerd => github.com/ukontainer/containerd v1.5.1-0.20211004093037-9c74b7617ea1
	github.com/containerd/containerd/api => github.com/containerd/containerd/api v0.0.0-20210908133427-99987f2a5eae
	github.com/docker/go-p9p => github.com/ukontainer/go-p9p v0.0.0-20211006131049-f1e80d0d54ed
)

require (
	github.com/containerd/console v1.0.3
	github.com/containerd/containerd v1.5.1
	github.com/containerd/containerd/api v0.0.0
	github.com/containerd/go-runc v1.0.0
	github.com/containerd/typeurl v1.0.2
	github.com/docker/go-p9p v0.0.0-20191112112554-37d97cf40d03
	github.com/gogo/protobuf v1.3.2
	github.com/hashicorp/go-multierror v1.1.0 // indirect
	github.com/opencontainers/runtime-spec v1.0.3-0.20210326190908-1c3f411f0417
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.8.1
	github.com/urfave/cli v1.22.2
	github.com/vishvananda/netlink v1.1.1-0.20201029203352-d40f9887b852
	github.com/vishvananda/netns v0.0.0-20200728191858-db3c7e526aae
	golang.org/x/sys v0.0.0-20210630005230-0f9fa26af87c
)
