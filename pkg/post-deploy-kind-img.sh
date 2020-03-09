#!/bin/sh

# XXX: need multi-arch image build
if [ $TRAVIS_ARCH != "amd64" ] || [ $TRAVIS_OS_NAME != "linux" ] ; then
    echo "This now only builds linux/amd64 image. Skipping"
    exit 0
fi

cp $TRAVIS_HOME/gopath/bin/${RUNU_PATH}runu k8s/
cd k8s
echo "$DOCKER_PASSWORD" | docker login -u "$DOCKER_USERNAME" --password-stdin
docker build -t thehajime/node-runu:v1.17.0 .
docker push thehajime/node-runu:v1.17.0
cd ..
