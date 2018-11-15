#!/bin/bash

mkdir -p /tmp/bundle/rootfs
mkdir -p /tmp/runu-root

fold_start() {
  echo -e "travis_fold:start:$1\033[33;1m$2\033[0m"
}

fold_end() {
  echo -e "\ntravis_fold:end:$1\r"
}


fold_start test.0 "preparation test"
# get script from moby
curl https://raw.githubusercontent.com/moby/moby/master/contrib/download-frozen-image-v2.sh \
     -o /tmp/download-frozen-image-v2.sh
bash /tmp/download-frozen-image-v2.sh /tmp/ thehajime/runu-base:$DOCKER_IMG_VERSION-$TRAVIS_OS_NAME

# extract images from layers
for layer in `find /tmp/ -name layer.tar`
do
 tar xvfz $layer -C /tmp/bundle/rootfs
done

rm -f config.json
runu spec

fold_end test.0

run_test()
{
    runu --debug --root=/tmp/runu-root run --bundle=/tmp/bundle foo
    sleep 5
    runu --debug --root=/tmp/runu-root kill foo 9
    runu --debug --root=/tmp/runu-root delete foo
}

# test hello-world
fold_start test.1 "test hello"
cat config.json | jq '.process.args |=["hello"] ' > /tmp/1
mv /tmp/1 config.json
cp config.json /tmp/bundle/
run_test
fold_end test.1

# test ping
fold_start test.2 "test ping"
cat config.json | jq '.process.args |=["ping","--","127.0.0.1"] ' > /tmp/1
mv /tmp/1 config.json
cp config.json /tmp/bundle/
run_test
fold_end test.2

# test python
fold_start test.3 "test python"
cat config.json | jq '.process.args |=["python", "imgs/python.iso", "imgs/python.img", "--", "-c", "print(\"hello world from python(runu)\")"] ' > /tmp/1
mv /tmp/1 config.json
cp config.json /tmp/bundle/
RUMP_VERBOSE=1 PYTHONHOME=/python run_test
fold_end test.3

#test nginx
fold_start test.4 "test nginx"
cat config.json | jq '.process.args |=["nginx", "imgs/data.iso", "tap:tap0", "config:lkl.json"] ' > /tmp/1
mv /tmp/1 config.json
cp config.json /tmp/bundle/
RUMP_VERBOSE=1 run_test
fold_end test.4
