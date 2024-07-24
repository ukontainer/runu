#!/bin/bash

. $(dirname "${BASH_SOURCE[0]}")/common.sh

DST_ADDR=$1

# XXX: need multi-arch image build
if [ $TRAVIS_ARCH != "amd64" ] || [ $TRAVIS_OS_NAME != "linux" ] ; then
    echo "This now only builds linux/amd64 image. Skipping"
    exit 0
fi

fold_start k8s.test.2 "k8s: kind setup"

# install runtime class
kubectl apply -f k8s/ukontainer-runtimeclass.yaml

fold_end k8s.test.2 ""

fold_start k8s.test.3 "k8s: hello world"
# install runu pod
## XXX: github action runners don't allow to pass ICMP at firewall
cat k8s/hello-world.yaml | sed "s/\$DOCKER_IMG_VERSION/$DOCKER_IMG_VERSION/" \
    | sed "s/8.8.8.8/$DST_ADDR/" \
    | kubectl apply -f -

kubectl get nodes -o wide -A
sleep 30
set -x
kubectl get pods -o wide -A
kubectl describe deployment/helloworld-runu
kubectl logs deployment/helloworld-runu |& tee /tmp/log.txt
grep "icmp_req=1" /tmp/log.txt

fold_end k8s.test.3 ""
