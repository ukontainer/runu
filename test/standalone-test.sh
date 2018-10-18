#!/bin/bash

mkdir -p /tmp/bundle/rootfs
mkdir -p /tmp/runu-root
docker export $(docker create thehajime/runu-base:$TRAVIS_OS_NAME sh) \
    | tar -C /tmp/bundle/rootfs -xvf -

rm -f config.json
./runu spec
cat config.json | jq '.process.args |=["ping","--","127.0.0.1"] ' > /tmp/1
mv /tmp/1 config.json
cp config.json /tmp/bundle/

./runu --debug --root=/tmp/runu-root run --bundle=/tmp/bundle foo
sleep 5
./runu --debug --root=/tmp/runu-root kill foo 9
./runu --debug --root=/tmp/runu-root delete foo

