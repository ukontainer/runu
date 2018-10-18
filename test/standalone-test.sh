#!/bin/bash

mkdir -p /tmp/bundle/rootfs
mkdir -p /tmp/runu-root

# get script from moby
curl -O https://raw.githubusercontent.com/moby/moby/master/contrib/download-frozen-image-v2.sh
bash download-frozen-image-v2.sh . thehajime/runu-base:$TRAVIS_OS_NAME

# extract images from layers
for layer in `find ./ -name layer.tar`
do
 tar xvfz $layer -C /tmp/bundle/rootfs
done

rm -f config.json
./runu spec
cat config.json | jq '.process.args |=["ping","--","127.0.0.1"] ' > /tmp/1
mv /tmp/1 config.json
cp config.json /tmp/bundle/

./runu --debug --root=/tmp/runu-root run --bundle=/tmp/bundle foo
sleep 5
./runu --debug --root=/tmp/runu-root kill foo 9
./runu --debug --root=/tmp/runu-root delete foo

