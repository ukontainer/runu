module github.com/ukontainer/runu

go 1.13

replace github.com/Sirupsen/logrus v1.4.2 => github.com/sirupsen/logrus v1.4.2

replace github.com/containerd/containerd v1.4.0 => github.com/ukontainer/containerd v1.2.1-0.20200911050324-9948b542f7f5

require (
	github.com/containerd/console v1.0.0
	github.com/containerd/containerd v1.4.0
	github.com/containerd/fifo v0.0.0-20200410184934-f15a3290365b // indirect
	github.com/containerd/go-runc v0.0.0-20200707131846-23d84c510c41
	github.com/containerd/ttrpc v1.0.1 // indirect
	github.com/containerd/typeurl v1.0.1
	github.com/docker/go-p9p v0.0.0-20191112112554-37d97cf40d03
	github.com/gogo/protobuf v1.3.1
	github.com/hashicorp/go-multierror v1.1.0 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/runtime-spec v1.0.2
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.4.2
	github.com/urfave/cli v1.22.1
	github.com/vishvananda/netlink v1.1.0
	github.com/vishvananda/netns v0.0.0-20200728191858-db3c7e526aae
	golang.org/x/sys v0.0.0-20200217220822-9197077df867
	google.golang.org/grpc v1.31.1 // indirect
)
