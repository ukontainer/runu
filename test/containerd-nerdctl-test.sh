#!/bin/bash

. $(dirname "${BASH_SOURCE[0]}")/common.sh

if [ $TRAVIS_OS_NAME != "osx" ] ; then
    echo "containerd and ctr runtime test only support with osx host. Skipped"
    exit 0
fi

CTR_GLOBAL_OPT="--debug -a /tmp/ctrd/run/containerd/containerd.sock --snapshotter=darwin"
NERDCTL_ARGS="--rm --env RUMP_VERBOSE=1"


###
### nerdctl tests
###

# preparation of nerdctl

# test hello-world (nerdctl)
fold_start test.nerdctl.1 "test hello (nerdctl)"
    sudo nerdctl $CTR_GLOBAL_OPT run $NERDCTL_ARGS \
        ${REGISTRY}ukontainer/runu-base:$DOCKER_IMG_VERSION hello
fold_end test.nerdctl.1

# test ping (nerdctl)
fold_start test.nerdctl.2 "test ping (nerdctl)"
    sudo nerdctl $CTR_GLOBAL_OPT run $NERDCTL_ARGS \
        --env LKL_ROOTFS=imgs/python.iso \
        ${REGISTRY}ukontainer/runu-base:$DOCKER_IMG_VERSION \
        ping -c5 127.0.0.1
fold_end test.nerdctl.2

# test python (nerdctl)
# XXX: PYTHONHASHSEED=1 is workaround for slow read of getrandom() on 4.19
# (4.16 doesn't have such)
fold_start test.nerdctl.3 "test python (nerdctl)"
    sudo nerdctl $CTR_GLOBAL_OPT run $NERDCTL_ARGS \
        --env HOME=/ --env PYTHONHOME=/python \
        --env LKL_ROOTFS=imgs/python.img \
        --env PYTHONHASHSEED=1 \
        ${REGISTRY}ukontainer/runu-base:$DOCKER_IMG_VERSION \
        python -c "print(\"hello world from python(docker-runu)\")"
fold_end test.nerdctl.3

# test nginx (nerdctl)
fold_start test.nerdctl.4 "test nginx (nerdctl)"
    sudo nerdctl $CTR_GLOBAL_OPT run $NERDCTL_ARGS \
        --env LKL_ROOTFS=imgs/data.iso \
        ${REGISTRY}ukontainer/runu-base:$DOCKER_IMG_VERSION \
        nginx &
sleep 3
sudo killall -9 nerdctl
fold_end test.nerdctl.4

# test alpine (nerdctl)
create_runu_aux_dir
fold_start test.nerdctl.5 "test alpine Linux on darwin (nerdctl)"
    sudo nerdctl $CTR_GLOBAL_OPT run $NERDCTL_ARGS \
        --env RUNU_AUX_DIR=$RUNU_AUX_DIR --env LKL_USE_9PFS=1 \
        library/alpine:latest /bin/busybox ls -l
fold_end test.nerdctl.5
