#!/bin/bash

. $(dirname "${BASH_SOURCE[0]}")/common.sh

DOCKER_RUN_ARGS="run --rm -i -e RUMP_VERBOSE=1  -e DEBUG=1 --runtime=runu-dev --net=none"

# prepare RUNU_AUX_DIR
create_runu_aux_dir

# update daemon.json
fold_start test.dockerd.0 "boot dockerd"
if [ $TRAVIS_OS_NAME = "linux" ] ; then

    (sudo cat /etc/docker/daemon.json 2>/dev/null || echo '{}') | \
        jq '.runtimes."runu-dev" |= {"path":"'${TRAVIS_HOME}'/gopath/bin/runu","runtimeArgs":[]}' | \
        tee /tmp/tmp.json
    sudo mv /tmp/tmp.json /etc/docker/daemon.json
    sudo service docker restart

elif [ $TRAVIS_OS_NAME = "osx" ] ; then

    sudo mkdir -p /etc/docker/
    git clone https://gist.github.com/aba357f73da4e14bc3f5cbeb00aeaea4.git /tmp/containerd-config-dockerd
    sudo cp /tmp/containerd-config-dockerd/daemon.json /etc/docker/

    # prepare dockerd
    mkdir -p /tmp/containerd-shim
    sudo killall containerd || true
    sudo dockerd --config-file /etc/docker/daemon.json > $HOME/dockerd.log 2>&1 &
    sleep 3
    sudo chmod 666 /tmp/var/run/docker.sock
    sudo chmod 777 /tmp/var/run/
    sudo ln -s /tmp/var/run/docker.sock /var/run/docker.sock

    # build docker (client)
    go get github.com/docker/cli/cmd/docker

    # XXX: docker run alpine stucks on macos: remove -d once it fixed
    DOCKER_RUN_EXT_ARGS="--platform=linux/amd64 -e LKL_USE_9PFS=1 --detach"
fi
fold_end test.dockerd.0 ""

# test hello-world
fold_start test.docker.0 "docker hello"
    docker $DOCKER_RUN_ARGS thehajime/runu-base:$DOCKER_IMG_VERSION hello
fold_end test.docker.0

# test ping
fold_start test.docker.1 "docker ping"
    docker $DOCKER_RUN_ARGS thehajime/runu-base:$DOCKER_IMG_VERSION \
           ping -c5 127.0.0.1
fold_end test.docker.1

# test python
# XXX: PYTHONHASHSEED=1 is workaround for slow read of getrandom() on 4.19
# (4.16 doesn't have such)
fold_start test.docker.2 "docker python"
    docker $DOCKER_RUN_ARGS -e HOME=/ \
           -e PYTHONHOME=/python -e LKL_ROOTFS=imgs/python.img \
           -e PYTHONHASHSEED=1 \
           thehajime/runu-base:$DOCKER_IMG_VERSION \
           python -c "print(\"hello world from python(docker-runu)\")"
fold_end test.docker.2

# test nginx
fold_start test.docker.3 "docker nginx"
CID=`docker $DOCKER_RUN_ARGS -d \
 -e LKL_ROOTFS=imgs/data.iso \
 thehajime/runu-base:$DOCKER_IMG_VERSION \
 nginx`
    sleep 2
    docker ps -a
    docker logs $CID
    docker kill $CID
fold_end test.docker.3


# alipine image test
fold_start test.docker.4 "docker alpine"
    docker $DOCKER_RUN_ARGS $DOCKER_RUN_EXT_ARGS \
           -e RUNU_AUX_DIR=$RUNU_AUX_DIR alpine uname -a

    docker $DOCKER_RUN_ARGS $DOCKER_RUN_EXT_ARGS \
           -e RUNU_AUX_DIR=$RUNU_AUX_DIR alpine ping -c 5 127.0.0.1

    docker $DOCKER_RUN_ARGS $DOCKER_RUN_EXT_ARGS \
           -e RUNU_AUX_DIR=$RUNU_AUX_DIR alpine dmesg | head

    docker $DOCKER_RUN_ARGS $DOCKER_RUN_EXT_ARGS \
           -e RUNU_AUX_DIR=$RUNU_AUX_DIR alpine ls -l /

    if [ $TRAVIS_OS_NAME = "linux" ] ; then
        # XXX: df -ha gives core dumps. remove this once fixed
        docker $DOCKER_RUN_ARGS $DOCKER_RUN_EXT_ARGS \
               -e RUNU_AUX_DIR=$RUNU_AUX_DIR alpine df -ha
    fi
fold_end test.docker.4
