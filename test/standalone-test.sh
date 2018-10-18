#!/bin/bash

mkdir -p /tmp/bundle/rootfs
mkdir -p /tmp/runu-root

# get script from moby
curl https://raw.githubusercontent.com/moby/moby/master/contrib/download-frozen-image-v2.sh \
     -o /tmp/download-frozen-image-v2.sh
bash /tmp/download-frozen-image-v2.sh /tmp/ thehajime/runu-base:$TRAVIS_OS_NAME

# extract images from layers
for layer in `find /tmp/ -name layer.tar`
do
 tar xvfz $layer -C /tmp/bundle/rootfs
done

rm -f config.json
./runu spec


run_test()
{
    ./runu --debug --root=/tmp/runu-root run --bundle=/tmp/bundle foo
    sleep 5
    ./runu --debug --root=/tmp/runu-root kill foo 9
    ./runu --debug --root=/tmp/runu-root delete foo
}

# test hello-world
cat config.json | jq '.process.args |=["hello"] ' > /tmp/1
mv /tmp/1 config.json
cp config.json /tmp/bundle/
run_test

# test ping
cat config.json | jq '.process.args |=["ping","--","127.0.0.1"] ' > /tmp/1
mv /tmp/1 config.json
cp config.json /tmp/bundle/
run_test

# test python
cat config.json | jq '.process.args |=["python", "imgs/python.iso", "imgs/python.img", "--", "-c", "print(\"hello world\")"] ' > /tmp/1
mv /tmp/1 config.json
cp config.json /tmp/bundle/
RUMP_VERBOSE=1 PYTHONHOME=/python run_test

