#!/bin/bash


fold_start() {
  echo -e "travis_fold:start:$1\033[33;1m$2\033[0m"
}

fold_end() {
  echo -e "\ntravis_fold:end:$1\r"
}


create_osx_chroot() {
    ROOTFS=$1
    if [ $TRAVIS_OS_NAME == "osx" ]; then
	sudo mount -t devfs devfs $ROOTFS/dev
    fi
}

create_runu_aux_dir() {
    # download pre-built frankenlibc
    curl -L https://dl.bintray.com/libos-nuse/x86_64-rumprun-linux/$TRAVIS_OS_NAME/frankenlibc.tar.gz \
	 -o /tmp/frankenlibc.tar.gz
    tar xfz /tmp/frankenlibc.tar.gz -C /tmp/
    cp /tmp/opt/rump/bin/rexec /tmp/rexec
    if [ -f /tmp/opt/rump/lib/libc.so ] ; then
	cp /tmp/opt/rump/lib/libc.so /tmp/libc.so
    fi

    export RUNU_AUX_DIR="/tmp"
}
