#!/bin/sh

# XXX: need multi-arch image build
if [ $TRAVIS_ARCH != "amd64" ] || [ $TRAVIS_OS_NAME != "linux" ] ; then
    echo "This now only builds linux/amd64 image. Skipping"
    exit 0
fi

docker push thehajime/node-runu:v1.17.0
