#!/bin/bash

if [ $TRAVIS_OS_NAME != "osx" ] ; then
    echo "dockerd and docker runtime test only support only with osx host. Skipped"
    exit 0
fi

. $(dirname "${BASH_SOURCE[0]}")/common.sh

DOCKER_ARGS="run --rm -e RUMP_VERBOSE=1  -e DEBUG=1 --runtime=runu-dev --net=none"

# prepare RUNU_AUX_DIR
create_runu_aux_dir

# build custom containerd
fold_start test.containerd.0 "containerd build"
HOMEBREW_NO_AUTO_UPDATE=1 brew install ukontainer/lkl/containerd
fold_end test.containerd.0 ""

#build custom dockerd
fold_start test.dockerd.0 "dockerd build"
HOMEBREW_NO_AUTO_UPDATE=1 brew install ukontainer/lkl/dockerd-darwin
fold_end test.dockerd.0 ""

# update daemon.json for dockerd
sudo mkdir -p /etc/docker/
sudo cp /tmp/containerd-config-dockerd/daemon.json /etc/docker/

# prepare dockerd
fold_start test.dockerd.0 "boot dockerd"
    sudo dockerd --config-file /etc/docker/daemon.json &
    sleep 3
    sudo chmod 666 /tmp/var/run/docker.sock
    sudo chmod 777 /tmp/var/run/
    sudo ln -s /tmp/var/run/docker.sock /var/run/docker.sock
fold_end test.dockerd.0 ""

# build docker (client)
go get github.com/docker/cli/cmd/docker

# test hello-world
fold_start test.dockerd.0 "docker hello"
    docker $DOCKER_ARGS thehajime/runu-base:0.1 hello
fold_end test.dockerd.0 ""

fold_start test.dockerd.1 "test ping"
    docker $DOCKER_ARGS thehajime/runu-base:0.1 ping -c 5 127.0.0.1
fold_end test.dockerd.1 ""
