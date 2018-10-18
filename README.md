# runu
OCI runtime for frankenlibc unikernel

# Installation

```
make
sudo cp runu /usr/local/bin/runu
```

add an entry to `/etc/docker/daemon.json`

```
        "runu": {
            "path": "/usr/local/bin/runu",
            "runtimeArgs": [
            ]
        },
```