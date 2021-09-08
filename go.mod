module github.com/ukontainer/runu

go 1.13

replace github.com/Sirupsen/logrus v1.4.2 => github.com/sirupsen/logrus v1.4.2

replace (
	github.com/containerd/containerd => github.com/ukontainer/containerd v1.5.1-0.20210908141113-5867ffe183c5
	github.com/containerd/containerd/api => github.com/containerd/containerd/api v0.0.0-20210908133427-99987f2a5eae
)

require (
	github.com/AdaLogics/go-fuzz-headers v0.0.0-20210715213245-6c3934b029d8 // indirect
	github.com/Microsoft/go-winio v0.5.0 // indirect
	github.com/Microsoft/hcsshim v0.8.21 // indirect
	github.com/containerd/console v1.0.2
	github.com/containerd/containerd v1.5.1
	github.com/containerd/containerd/api v0.0.0
	github.com/containerd/go-runc v1.0.0
	github.com/containerd/typeurl v1.0.2
	github.com/docker/go-p9p v0.0.0-20191112112554-37d97cf40d03
	github.com/gogo/protobuf v1.3.2
	github.com/hashicorp/go-multierror v1.1.0 // indirect
	github.com/imdario/mergo v0.3.12 // indirect
	github.com/moby/sys/signal v0.5.1-0.20210723232958-8a51b5cc8879 // indirect
	github.com/opencontainers/runc v1.0.2 // indirect
	github.com/opencontainers/runtime-spec v1.0.3-0.20210326190908-1c3f411f0417
	github.com/pelletier/go-toml v1.9.3 // indirect
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.8.1
	github.com/urfave/cli v1.22.2
	github.com/vishvananda/netlink v1.1.1-0.20201029203352-d40f9887b852
	github.com/vishvananda/netns v0.0.0-20200728191858-db3c7e526aae
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.21.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.0.0-RC1 // indirect
	golang.org/x/sys v0.0.0-20210630005230-0f9fa26af87c
	k8s.io/apiserver v0.22.0 // indirect
	k8s.io/cri-api v0.22.0 // indirect
)
