#!/bin/sh

cp runu k8s
cd k8s
echo "$DOCKER_PASSWORD" | docker login -u "$DOCKER_USERNAME" --password-stdin
docker build -t thehajime/node-runu:v1.17.0 .
docker push thehajime/node-runu:v1.17.0
cd ..
