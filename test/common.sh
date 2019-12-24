#!/bin/bash


fold_start() {
  echo -e "travis_fold:start:$1\033[33;1m$2\033[0m"
  travis_time_start
}

fold_end() {
  travis_time_finish
  echo -e "\ntravis_fold:end:$1\r"
}


create_osx_chroot() {
    ROOTFS=$1
    if [ $TRAVIS_OS_NAME == "osx" ]; then
	sudo mount -t devfs devfs $ROOTFS/dev
    fi
}

create_runu_aux_dir() {
    export RUNU_AUX_DIR="/tmp/runu"

    if [ -a $RUNU_AUX_DIR/lkick ] ; then
	return
    fi

    # download pre-built frankenlibc
    curl -L https://dl.bintray.com/ukontainer/ukontainer/$TRAVIS_OS_NAME/$ARCH/frankenlibc.tar.gz \
	 -o /tmp/frankenlibc.tar.gz
    tar xfz /tmp/frankenlibc.tar.gz -C /tmp/
    cp /tmp/opt/rump/bin/rexec $RUNU_AUX_DIR/rexec
    cp /tmp/opt/rump/bin/lkick $RUNU_AUX_DIR/lkick
    if [ $TRAVIS_OS_NAME == "osx" ]; then
	curl -L https://dl.bintray.com/ukontainer/ukontainer/linux/amd64/frankenlibc.tar.gz \
	     -o /tmp/frankenlibc-linux.tar.gz
	tar xfz /tmp/frankenlibc-linux.tar.gz -C /tmp opt/rump/lib/libc.so
    fi
    if [ -f /tmp/opt/rump/lib/libc.so ] ; then
	cp /tmp/opt/rump/lib/libc.so $RUNU_AUX_DIR/libc.so
    fi
}

# common variables
OSNAME=$(uname -s)
if [ -z $TRAVIS_OS_NAME ] ; then
    if [ $OSNAME = "Linux" ] ; then
	TRAVIS_OS_NAME="linux"
    elif [ $OSNAME = "Darwin" ] ; then
	TRAVIS_OS_NAME="osx"
    fi
fi

PNAME=$(uname -m)
if [ -z $ARCH ] ; then
    if [ $PNAME = "x86_64" ] ; then
	ARCH="amd64"
    elif [ $PNAME = "aarch64" ] ; then
	ARCH="arm"
    elif [ $PNAME = "armv7l" ] ; then
	ARCH="arm"
    fi
fi

DOCKER_IMG_VERSION=0.2
