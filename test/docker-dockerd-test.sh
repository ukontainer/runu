#!/bin/bash

if [ $TRAVIS_OS_NAME != "osx" ] ; then
    echo "dockerd and docker runtime test only support only with osx host. Skipped"
    exit 0
fi

. $(dirname "${BASH_SOURCE[0]}")/common.sh

DOCKER_ARGS="--rm -e UMP_VERBOSE=1  -e DEBUG=1 --runtime=runu-dev --net=none"

# prepare RUNU_AUX_DIR
create_runu_aux_dir

# update daemon.json
(sudo cat /etc/docker/daemon.json 2>/dev/null || echo '{}') | \
    jq '.runtimes.runu |= {"path":"'${TRAVIS_HOME}'/gopath/bin/runu","runtimeArgs":[]}' | \
    tee /tmp/tmp.json
sudo mv /tmp/tmp.json /etc/docker/daemon.json
sudo service docker restart

# build custom containerd
fold_start test.containerd.0 "containerd build"
HOMEBREW_NO_AUTO_UPDATE=1 brew install libos-nuse/lkl/containerd
fold_end test.containerd.0 ""

# prepare containerd
fold_start test.containerd.0 "boot containerd"
    git clone https://gist.github.com/aba357f73da4e14bc3f5cbeb00aeaea4.git /tmp/containerd-config
    cp /tmp/containerd-config/config.toml /tmp/
    sed "s/501/$UID/" /tmp/config.toml > /tmp/a
    mv /tmp/a /tmp/config.toml

    mkdir /tmp/containerd-shim
    sudo killall containerd || true
    containerd -l debug -c /tmp/config.toml &
    sleep 3
    killall containerd
    sudo containerd -l debug -c /tmp/config.toml &
    sleep 3
    chmod 755 /tmp/ctrd
    ls -lRa /tmp/ctrd
fold_end test.containerd.0 ""

#build custom dockerd
fold_start test.dockerd.0 "dockerd build"
HOMEBREW_NO_AUTO_UPDATE=1 brew install libos-nuse/lkl/dockerd-darwin
fold_end test.dockerd.0 ""

# prepare dockerd
fold_start test.dockerd.0 "boot dockerd"
    dockerd --config-file /etc/docker/daemon.json --experimental -l debug
fold_end test.dockerd.0 ""

# test hello-world
fold_start test.dockerd.0 "docker hello"
    docker $DOCKER_ARGS --platform=linux/amd64 thehajime/runu-base:0.1 hello
fold_end test.dockerd.0 ""

fold_start test.dockerd.1 "test ping"
    docker $DOCKER_ARGS --platform=linux/amd64 thehajime/runu-base:0.1 ping 127.0.0.1
fold_end test.dockerd.1 ""
