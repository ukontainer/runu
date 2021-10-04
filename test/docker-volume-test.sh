#!/bin/bash

. $(dirname "${BASH_SOURCE[0]}")/common.sh

# prepare RUNU_AUX_DIR
create_runu_aux_dir

DOCKER_RUN_ARGS="run --rm -i --runtime=runu-dev --net=none $DOCKER_RUN_ARGS_ARCH"

if ! [ $TRAVIS_OS_NAME = "osx" ] ; then
fold_start test.docker.0 "docker-volume: -v option should fail when specified with -e LKL_ROOTFS"

MNT_SRC=$PWD/host_dir
MNT_DST=/mnt
mkdir -p $MNT_SRC
docker $DOCKER_RUN_ARGS -e RUMP_VERBOSE=1 \
       -e LKL_ROOTFS=imgs/python.img \
       -v $MNT_SRC:$MNT_DST \
       ${REGISTRY}ukontainer/runu-base:$DOCKER_IMG_VERSION \
       hello  2> fail_log || true
cat fail_log
cat fail_log | grep "OCI runtime create failed" >/dev/null
rm -rf $MNT_SRC fail_log

fold_end test.docker.0
fi

if [ $TRAVIS_OS_NAME = "linux" ] ; then

fold_start test.docker.1 "docker-volume: naive -v option for directory"

MNT_SRC=$PWD/host_dir
MNT_DST=/mnt
mkdir -p $MNT_SRC
touch    $MNT_SRC/foo
touch    $MNT_SRC/bar
docker $DOCKER_RUN_ARGS -e RUMP_VERBOSE=1 \
       -v $MNT_SRC:$MNT_DST \
       ${REGISTRY}ukontainer/runu-python:$DOCKER_IMG_VERSION \
       python -c "import os; print(os.listdir('/mnt'))" | egrep "foo.*bar|bar.*foo"
rm -rf $MNT_SRC

fold_end test.docker.1

fold_start test.docker.2 "docker-volume: naive -v option for file"

MNT_SRC=/tmp/hello.txt
MNT_DST=/mnt/hello.txt
echo "hello_world" > $MNT_SRC
docker $DOCKER_RUN_ARGS -e RUMP_VERBOSE=1 \
       -v $MNT_SRC:$MNT_DST \
       -e RUNU_AUX_DIR=$RUNU_AUX_DIR alpine /bin/busybox cat /mnt/hello.txt | grep 'hello_world'

rm -f $MNT_SRC
fold_end test.docker.2

fold_start test.docker.3 "docker-volume: naive -v option for named volume"

docker run --rm -v named_vol:/mnt alpine touch /mnt/foo /mnt/bar
docker $DOCKER_RUN_ARGS -e RUMP_VERBOSE=1 \
       -v named_vol:/mnt \
       ${REGISTRY}ukontainer/runu-python:$DOCKER_IMG_VERSION \
       python -c "import os; print(os.listdir('/mnt'))" | egrep "foo.*bar|bar.*foo"
docker volume rm named_vol
fold_end test.docker.3

fold_start test.docker.4 "docker-volume: read only -v option for directory"


MNT_SRC=$PWD/host_dir
MNT_DST=/mnt
mkdir -p $MNT_SRC
touch    $MNT_SRC/foo
touch    $MNT_SRC/bar
docker $DOCKER_RUN_ARGS -e RUMP_VERBOSE=1 \
       -v $MNT_SRC:$MNT_DST:ro \
       ${REGISTRY}ukontainer/runu-python:$DOCKER_IMG_VERSION \
       python -c "import os; print(os.listdir('/mnt'))" | egrep "foo.*bar|bar.*foo"

docker $DOCKER_RUN_ARGS -e RUMP_VERBOSE=1 \
       -v $MNT_SRC:$MNT_DST:ro \
       ${REGISTRY}ukontainer/runu-python:$DOCKER_IMG_VERSION \
       python -c "f=open('${MNT_DST}/hello.txt', 'w'), print('hello',file=f);close(f)" 2> fail_log || true
cat fail_log
cat fail_log | grep "OSError" >/dev/null
rm -rf $MNT_SRC

2> fail_log || true

fold_end test.docker.4

fold_start test.docker.5 "docker-volume: read only -v option for file"
MNT_SRC=/tmp/hello.txt
MNT_DST=/mnt/hello.txt
echo "hello_world" > $MNT_SRC
docker $DOCKER_RUN_ARGS -e RUMP_VERBOSE=1 \
       -v $MNT_SRC:$MNT_DST:ro \
       -e RUNU_AUX_DIR=$RUNU_AUX_DIR alpine /bin/busybox cat $MNT_DST | grep 'hello_world'

docker $DOCKER_RUN_ARGS -e RUMP_VERBOSE=1 \
       -v $MNT_SRC:$MNT_DST:ro \
       ${REGISTRY}ukontainer/runu-python:$DOCKER_IMG_VERSION \
       python -c "f=open('${MNT_DST}', 'w'), print('hello',file=f);close(f)" 2> fail_log || true
cat fail_log
cat fail_log | grep "OSError" >/dev/null
rm -f $MNT_SRC
fold_end test.docker.5

fold_start test.docker.6 "docker-volume: read only -v option for named volume"

docker run --rm -v named_vol:/mnt alpine touch /mnt/foo /mnt/bar
docker $DOCKER_RUN_ARGS -e RUMP_VERBOSE=1 \
       -v named_vol:/mnt:ro \
       ${REGISTRY}ukontainer/runu-python:$DOCKER_IMG_VERSION \
       python -c "import os; print(os.listdir('/mnt'))" | egrep "foo.*bar|bar.*foo"

docker $DOCKER_RUN_ARGS -e RUMP_VERBOSE=1 \
       -v named_vol:/mnt:ro \
       ${REGISTRY}ukontainer/runu-python:$DOCKER_IMG_VERSION \
       python -c "f=open('/mnt/hello.txt', 'w'), print('hello',file=f);close(f)" 2> fail_log || true
cat fail_log
cat fail_log | grep "OSError" >/dev/null

docker volume rm named_vol

fold_end test.docker.6
fi
