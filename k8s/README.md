## Container runtimeClass for uKontainer

uKontainer (runu) supports runtimeClass to use alternate runtime
instead of runc over CRI mechanism. We only support (and tested)
containerd CRI; thus the default dockershim need to be stopped and
containerd should be used instead.


### Installation

1. containerd configuration

/etc/containerd/config.toml
```
    [plugins.cri.containerd]
(snip)
      [plugins.cri.containerd.default_runtime]
        runtime_type = "io.containerd.runtime.v1.linux"
        runtime_engine = "/usr/bin/runu"
        runtime_root = ""
      [plugins.cri.containerd.runtimes.runu]
        runtime_type = "io.containerd.runtime.v1.linux"

```

The last two lines are added one. Let's restart containerd afterward
by `systemctl restart containerd`.


2. Installing RuntimeClass resoure

Need to install a runtime class resource by:

```
kubectl apply -f ./k8s/ukontainer-runtimeclass.yaml
```

3. Install runu binary

See the instruction described at https://bintray.com/ukontainer/debian.
Alternatively, you can use a KinD (k8s in Docker) image which contains runu
binary (https://hub.docker.com/r/thehajime/node-runu).

### Usage



The pod configuration (.yaml file) should look like the following:


```
    spec:
      runtimeClassName: ukontainer
      containers:
        - name: test-pod
          image: test-pod:1.0
          imagePullPolicy: Always
          ports:
          - containerPort: 8080
```

### Reference

https://kubernetes.io/docs/concepts/containers/runtime-class/