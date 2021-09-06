#!/bin/bash

travis_nanoseconds() {
  local cmd='date'
  local format='+%s%N'

  if hash gdate >/dev/null 2>&1; then
    cmd='gdate'
  elif [[ "${TRAVIS_OS_NAME}" == osx ]]; then
    format='+%s000000000'
  fi

  "${cmd}" -u "${format}"
}

travis_time_start() {
  TRAVIS_TIMER_ID="$(printf %08x $((RANDOM * RANDOM)))"
  TRAVIS_TIMER_START_TIME="$(travis_nanoseconds)"
  export TRAVIS_TIMER_ID TRAVIS_TIMER_START_TIME
  echo -en "travis_time:start:$TRAVIS_TIMER_ID\\r${ANSI_CLEAR}"
}


travis_time_finish() {
  local result="${?}"
  local travis_timer_end_time
  local event="${1}"
  travis_timer_end_time="$(travis_nanoseconds)"
  local duration
  duration="$((travis_timer_end_time - TRAVIS_TIMER_START_TIME))"
  echo -en "travis_time:end:${TRAVIS_TIMER_ID}:start=${TRAVIS_TIMER_START_TIME},finish=${travis_timer_end_time},duration=${duration},event=${event}\\r${ANSI_CLEAR}"
  return "${result}"
}

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
    URL="https://github.com/ukontainer/frankenlibc/releases/download/dev/frankenlibc-$ARCH-$TRAVIS_OS_NAME.tar.gz"
    URL_LINUX="https://github.com/ukontainer/frankenlibc/releases/download/dev/frankenlibc-amd64-linux.tar.gz"
    curl -L $URL -o /tmp/frankenlibc.tar.gz
    tar xfz /tmp/frankenlibc.tar.gz -C /tmp/
    cp /tmp/opt/rump/bin/rexec $RUNU_AUX_DIR/rexec
    cp /tmp/opt/rump/bin/lkick $RUNU_AUX_DIR/lkick
    if [ $TRAVIS_OS_NAME == "osx" ]; then
	curl -L $URL_LINUX -o /tmp/frankenlibc-linux.tar.gz
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
    elif [ $DEB_ARCH = "arm64" ] ; then
	ARCH="arm64"
    elif [ $DEB_ARCH = "armhf" ] ; then
	ARCH="arm"
    fi
fi

DOCKER_IMG_VERSION=0.5
