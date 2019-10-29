#!/bin/sh

sudo apt-get install ./$PACKAGE_NAME_VERSION
DOCKER_RUN_ARGS="run --rm -i -e RUMP_VERBOSE=1 -e DEBUG=1 --runtime=runu-dev --net=none $DOCKER_RUN_ARGS_ARCH"

docker $DOCKER_RUN_ARGS $DOCKER_RUN_EXT_ARGS alpine uname -a
