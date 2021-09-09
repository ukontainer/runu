#!/bin/bash

. $(dirname "${BASH_SOURCE[0]}")/common.sh

if [ $TRAVIS_OS_NAME != "osx" ] ; then
    echo "containerd and ctr runtime test only support with osx host. Skipped"
    exit 0
fi

CTR_ARGS="--rm --runtime=io.containerd.runu.v1 --fifo-dir /tmp/ctrd --env RUMP_VERBOSE=1"
CTR_GLOBAL_OPT="--debug -a /tmp/ctrd/run/containerd/containerd.sock"

sudo rm -rf /tmp/ctrd/

# prepare containerd
fold_start test.containerd.0 "boot containerd"
    git clone https://gist.github.com/aba357f73da4e14bc3f5cbeb00aeaea4.git \
	/tmp/containerd-config || true
    cp /tmp/containerd-config/config.toml /tmp/
    sed "s/501/$UID/" /tmp/config.toml > /tmp/a
    mv /tmp/a /tmp/config.toml

    mkdir -p /tmp/containerd-shim
    sudo killall containerd || true
    containerd -l debug -c /tmp/config.toml &
    sleep 3
    killall containerd
    sudo containerd -l debug -c /tmp/config.toml > /tmp/containerd.log 2>&1 &
    sleep 3
    chmod 755 /tmp/ctrd

    ctr $CTR_GLOBAL_OPT version
fold_end test.containerd.0 ""


# pull an image
fold_start test.containerd.0 "pull image"
    ctr -a /tmp/ctrd/run/containerd/containerd.sock i pull \
       docker.io/ukontainer/runu-base:$DOCKER_IMG_VERSION
    ctr -a /tmp/ctrd/run/containerd/containerd.sock i pull \
        --platform=linux/amd64 docker.io/library/alpine:latest
fold_end test.containerd.0 "pull image"

# test hello-world
fold_start test.containerd.1 "test hello"
    ctr $CTR_GLOBAL_OPT run $CTR_ARGS \
        docker.io/ukontainer/runu-base:$DOCKER_IMG_VERSION hello hello
fold_end test.containerd.1

# test ping
fold_start test.containerd.2 "test ping"
    ctr $CTR_GLOBAL_OPT run $CTR_ARGS \
        --env LKL_ROOTFS=imgs/python.iso \
        docker.io/ukontainer/runu-base:$DOCKER_IMG_VERSION hello \
        ping -c5 127.0.0.1
fold_end test.containerd.2

# test python
# XXX: PYTHONHASHSEED=1 is workaround for slow read of getrandom() on 4.19
# (4.16 doesn't have such)
fold_start test.containerd.3 "test python"
    ctr $CTR_GLOBAL_OPT run $CTR_ARGS \
        --env HOME=/ --env PYTHONHOME=/python \
        --env LKL_ROOTFS=imgs/python.img \
        --env PYTHONHASHSEED=1 \
        docker.io/ukontainer/runu-base:$DOCKER_IMG_VERSION hello \
        python -c "print(\"hello world from python(docker-runu)\")"
fold_end test.containerd.3

# test nginx
fold_start test.containerd.4 "test nginx"
    ctr $CTR_GLOBAL_OPT run $CTR_ARGS \
        --env LKL_ROOTFS=imgs/data.iso \
        docker.io/ukontainer/runu-base:$DOCKER_IMG_VERSION hello \
        nginx &
sleep 3
killall -9 ctr
fold_end test.containerd.4

# test alpine
# prepare RUNU_AUX_DIR
create_runu_aux_dir

fold_start test.containerd.5 "test alpine Linux on darwin"
    ctr $CTR_GLOBAL_OPT run $CTR_ARGS \
        --env RUNU_AUX_DIR=$RUNU_AUX_DIR --env LKL_USE_9PFS=1 \
        docker.io/library/alpine:latest alpine1 /bin/busybox ls -l
fold_end test.containerd.5
