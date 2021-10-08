[![CI](https://github.com/ukontainer/runu/actions/workflows/ci.yml/badge.svg)](https://github.com/ukontainer/runu/actions/workflows/ci.yml)
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

Optionally, you can install debian package from the repository.

```
# register apt repository
curl -s https://packagecloud.io/install/repositories/ukontainer/runu/script.deb.sh | sudo bash
# install the package
sudo apt-get install docker-runu
```