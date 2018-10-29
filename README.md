[![Build Status](https://travis-ci.org/libos-nuse/runu.svg?branch=master)](https://travis-ci.org/libos-nuse/runu)
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
