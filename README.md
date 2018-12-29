[![Build Status](https://travis-ci.org/ukontainer/runu.svg?branch=master)](https://travis-ci.org/ukontainer/runu)
[![Go Report Card](https://goreportcard.com/badge/github.com/libos-nuse/runu)](https://goreportcard.com/report/github.com/libos-nuse/runu)


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
