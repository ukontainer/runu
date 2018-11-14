#!/bin/bash

IMG_VERSION=0.1

if [ $TRAVIS_OS_NAME != "linux" ] ; then
    echo "docker OCI runtime test only support with Linux host. Skipped"
    exit 0
fi


fold_start() {
  echo -e "travis_fold:start:$1\033[33;1m$2\033[0m"
}

fold_end() {
  echo -e "\ntravis_fold:end:$1\r"
}

# update daemon.json
(sudo cat /etc/docker/daemon.json 2>/dev/null || echo '{}') | \
    jq '.runtimes.runu |= {"path":"'${TRAVIS_HOME}'/gopath/bin/runu","runtimeArgs":[]}' | \
    tee /tmp/tmp.json
sudo mv /tmp/tmp.json /etc/docker/daemon.json
sudo service docker restart


# test hello-world
fold_start test.docker.0 "test hello"
docker run --rm -i --runtime=runu thehajime/runu-base:$IMG_VERSION hello
fold_end test.docker.0

# test ping
fold_start test.docker.1 "test ping"
docker run --rm -i -e RUMP_VERBOSE=1 -e LKL_OFFLOAD=1 \
 --runtime=runu thehajime/runu-base:$IMG_VERSION \
 ping imgs/python.iso -- -c5 127.0.0.1
fold_end test.docker.1

# test python
fold_start test.docker.2 "test python"
docker run --rm -i -e RUMP_VERBOSE=1 -e LKL_OFFLOAD=1 \
 -e HOME=/ -e PYTHONHOME=/python \
 --runtime=runu thehajime/runu-base:$IMG_VERSION \
 python imgs/python.iso imgs/python.img -- \
 -c "print(\"hello world from python(docker-runu)\")"
fold_end test.docker.2

# test nginx
fold_start test.docker.3 "test nginx"
CID=`docker run -d --rm -i -e RUMP_VERBOSE=1 -e LKL_OFFLOAD=1 \
 --runtime=runu thehajime/runu-base:$IMG_VERSION \
 nginx imgs/python.iso imgs/data.iso`
sleep 2
docker logs $CID
docker kill $CID
fold_end test.docker.3

