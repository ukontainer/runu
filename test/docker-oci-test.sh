#!/bin/bash

if [ $TRAVIS_OS_NAME != "linux" ] ; then
    echo "docker OCI runtime test only support with Linux host. Skipped"
    exit 0
fi

. $(dirname "${BASH_SOURCE[0]}")/common.sh

# prepare RUNU_AUX_DIR
create_runu_aux_dir

# update daemon.json
(sudo cat /etc/docker/daemon.json 2>/dev/null || echo '{}') | \
    jq '.runtimes.runu |= {"path":"'${TRAVIS_HOME}'/gopath/bin/runu","runtimeArgs":[]}' | \
    tee /tmp/tmp.json
sudo mv /tmp/tmp.json /etc/docker/daemon.json
sudo service docker restart

# test hello-world
fold_start test.docker.0 "docker hello"
docker run --rm -i --runtime=runu thehajime/runu-base:$DOCKER_IMG_VERSION hello
fold_end test.docker.0

# test ping
fold_start test.docker.1 "docker ping"
docker run --rm -i -e RUMP_VERBOSE=1 -e LKL_OFFLOAD=1 -e LKL_ROOTFS=imgs/python.iso \
 --runtime=runu thehajime/runu-base:$DOCKER_IMG_VERSION \
 ping -c5 127.0.0.1
fold_end test.docker.1

# test python
fold_start test.docker.2 "docker python"
docker run --rm -i -e RUMP_VERBOSE=1 -e LKL_OFFLOAD=1 \
 -e HOME=/ -e PYTHONHOME=/python -e LKL_ROOTFS=imgs/python.img \
 --runtime=runu thehajime/runu-base:$DOCKER_IMG_VERSION \
 python -c "print(\"hello world from python(docker-runu)\")"
fold_end test.docker.2

# test nginx
fold_start test.docker.3 "docker nginx"
CID=`docker run -d --rm -i -e RUMP_VERBOSE=1 \
 -e LKL_ROOTFS=imgs/data.iso \
 --runtime=runu thehajime/runu-base:$DOCKER_IMG_VERSION \
 nginx`
sleep 2
docker ps -a
docker logs $CID
docker kill $CID
fold_end test.docker.3

# alipine image test
fold_start test.docker.4 "docker alpine ldso"
docker run -i --runtime=runu --rm -e RUMP_VERBOSE=1 \
       -e RUNU_AUX_DIR=$RUNU_AUX_DIR alpine uname -a
docker run -i --runtime=runu --rm -e UMP_VERBOSE=1 \
       -e RUNU_AUX_DIR=$RUNU_AUX_DIR alpine ping -c 5 127.0.0.1
docker run -i --runtime=runu --rm -e UMP_VERBOSE=1 \
       -e RUNU_AUX_DIR=$RUNU_AUX_DIR alpine dmesg | head
docker run -i --runtime=runu --rm -e UMP_VERBOSE=1 \
       -e RUNU_AUX_DIR=$RUNU_AUX_DIR alpine ls -l /
docker run -i --runtime=runu --rm -e UMP_VERBOSE=1 \
       -e RUNU_AUX_DIR=$RUNU_AUX_DIR alpine df -ha
fold_end test.docker.4
