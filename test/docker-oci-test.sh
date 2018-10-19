#!/bin/bash


if [ $TRAVIS_OS_NAME != "linux" ] ; then
    echo "docker OCI runtime test only support with Linux host. Skipped"
    exit 0
fi

# update daemon.json
(sudo cat /etc/docker/daemon.json 2>/dev/null || echo '{}') | \
    jq '.runtimes.runu |= {"path":"/usr/local/bin/runu","runtimeArgs":[]}' | \
    tee /tmp/tmp.json
sudo mv /tmp/tmp.json /etc/docker/daemon.json

# test hello-world
docker run --rm -i --runtime=runu thehajime/runu-base:linux hello

# test ping
docker run --rm -i -e RUMP_VERBOSE=1 -e LKL_OFFLOAD=1 \
 --runtime=runu thehajime/runu-base:linux \
 ping imgs/python.iso -- 127.0.0.1

# test python
docker run --rm -i -e RUMP_VERBOSE=1 -e LKL_OFFLOAD=1 \
 -e HOME=/ -e PYTHONHOME=/python \
 --runtime=runu thehajime/runu-base:linux \
 python imgs/python.iso imgs/python.img -- \
 -c "print(\"hello world from python(runu)\")"

# test nginx
docker run --rm -i -e RUMP_VERBOSE=1 -e LKL_OFFLOAD=1 \
 -e HOME=/ -e PYTHONHOME=/python \
 --runtime=runu thehajime/runu-base:linux \
 nginx imgs/data.iso

