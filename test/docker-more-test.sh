#!/bin/bash

. $(dirname "${BASH_SOURCE[0]}")/common.sh

if [ $TRAVIS_ARCH != "amd64" ] || [ $TRAVIS_OS_NAME != "linux" ] ; then
    echo "those tests contain tap device creation, which is only supported on amd64/linux instance. Skipped"
    exit 0
fi

# prepare RUNU_AUX_DIR
create_runu_aux_dir

DOCKER_RUN_ARGS="run --rm -i --runtime=runu-dev --net=none $DOCKER_RUN_ARGS_ARCH"

# 0. tap config
# 1. local json / ip addr config
# 2. local json / no ip addr config
# 3. no json
# 4. image json
# and more..

# 0. tap config
TAP_IFNAME=tap0
set +e
set -x
TAP_EXIST=`ip link |grep $TAP_IFNAME`
set -e

if [ -z "$TAP_EXIST" ] ; then
    sudo ip tuntap add $TAP_IFNAME mode tap user $USER
    sudo ifconfig $TAP_IFNAME up
    sudo brctl addif docker0 $TAP_IFNAME

    ip addr
fi
set +x

# 1. local json / ip addr config
fold_start test.docker.conf.1 "docker+: local json / ip addr config"
cat > /tmp/lkl.json <<EOF
{
    "gateway": "172.17.0.1",
    "interfaces": [
        {
            "ip": "172.17.0.50",
            "masklen": "24",
            "name": "$TAP_IFNAME",
            "type": "rumpfd"
        }
    ]
}
EOF
    cat /tmp/lkl.json
    docker $DOCKER_RUN_ARGS $DOCKER_RUN_EXT_ARGS \
           -e RUNU_AUX_DIR=$RUNU_AUX_DIR \
	   -e LKL_NET=$TAP_IFNAME -e LKL_CONFIG=/tmp/lkl.json \
           alpine $ALPINE_PREFIX ip addr | grep 172.17.0.50
fold_end test.docker.conf.1

# 2. local json / no ip addr config
fold_start test.docker.conf.2 "docker+: local json / no ip addr config"
cat > /tmp/lkl.json <<EOF
{
    "interfaces": [
        {
            "name": "$TAP_IFNAME",
            "type": "rumpfd"
        }
    ],
	"debug": "1"
}
EOF
    docker $DOCKER_RUN_ARGS $DOCKER_RUN_EXT_ARGS \
           -e RUNU_AUX_DIR=$RUNU_AUX_DIR \
	   -e LKL_NET=$TAP_IFNAME -e LKL_CONFIG=/tmp/lkl.json \
           alpine $ALPINE_PREFIX ip addr
fold_end test.docker.conf.2

# 3. no json
fold_start test.docker.conf.3 "docker+: no json"
    docker $DOCKER_RUN_ARGS $DOCKER_RUN_EXT_ARGS \
           -e RUNU_AUX_DIR=$RUNU_AUX_DIR \
	   -e LKL_NET=$TAP_IFNAME \
           alpine $ALPINE_PREFIX ip addr
fold_end test.docker.conf.3

# 4. image json
fold_start test.docker.conf.4 "docker+: image json"
    docker $DOCKER_RUN_ARGS -e RUMP_VERBOSE=1 \
           -e LKL_NET=$TAP_IFNAME -e LKL_CONFIG=/lkl.json \
           ukontainer/runu-base:$DOCKER_IMG_VERSION \
           hello | grep virtio-mmio.0
fold_end test.docker.conf.4


# test named
fold_start test.docker.5 "docker+: named"
cat > /tmp/lkl.json <<EOF
{
    "gateway": "172.17.0.1",
    "interfaces": [
        {
            "ip": "172.17.0.50",
            "masklen": "24",
            "ipv6": "2001:db8::2",
            "masklen6": "64",
            "name": "$TAP_IFNAME",
            "type": "rumpfd"
        }
    ]
}
EOF

    CID=named-docker
    docker $DOCKER_RUN_ARGS -d --name $CID \
     -e LKL_NET=$TAP_IFNAME\
     -e LKL_CONFIG=/tmp/lkl.json\
     ukontainer/runu-named:$DOCKER_IMG_VERSION \
     named -c /etc/bind/named.conf -g

    sleep 10
    docker ps -a
    docker logs $CID
    ping -c 1 172.17.0.50 || true
    ping6 -c 1 2001:db8::2 || true
    nslookup ns.hoge.local 172.17.0.50
    nslookup host.hoge.local 172.17.0.50
    nslookup host.hoge.local 2001:db8::2
    docker kill $CID
fold_end test.docker.5
