#!/bin/bash

. $(dirname "${BASH_SOURCE[0]}")/common.sh

KIND_VERSION=v0.11.1
KIND_IMG_VERSION=v1.21.1

# XXX: need multi-arch image build
if [ $TRAVIS_ARCH != "amd64" ] || [ $TRAVIS_OS_NAME != "linux" ] ; then
    echo "This now only builds linux/amd64 image. Skipping"
    exit 0
fi

# Kind preparation
fold_start k8s.test.0 "k8s: KinD image preparation"

# prepare RUNU_AUX_DIR
create_runu_aux_dir
cp $RUNU_AUX_DIR/libc.so k8s/
cp $RUNU_AUX_DIR/lkick k8s/

# Build kind node docker image
cp $TRAVIS_HOME/gopath/bin/${RUNU_PATH}runu k8s/
cd k8s
echo "$DOCKER_PASSWORD" | docker login -u "$DOCKER_USERNAME" --password-stdin
docker build -t thehajime/node-runu:$KIND_IMG_VERSION .
cd ..

fold_end k8s.test.0 ""

fold_start k8s.test.1 "k8s: install kind cli"

GO111MODULE="on" go get sigs.k8s.io/kind@$KIND_VERSION

# install kubectl
curl -LO \
     https://storage.googleapis.com/kubernetes-release/release/$(curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt)/bin/linux/amd64/kubectl
chmod +x kubectl
sudo mv kubectl /usr/local/bin/

fold_end k8s.test.1 ""


fold_start k8s.test.2 "k8s: kind setup"
kind create cluster --image thehajime/node-runu:$KIND_IMG_VERSION --config k8s/kind-cluster.yaml
kubectl get pods -o wide -A
kubectl get nodes -o wide -A

sleep 60

# install runtime class
kubectl apply -f k8s/ukontainer-runtimeclass.yaml

fold_end k8s.test.2 ""

fold_start k8s.test.3 "k8s: hello world"
# install runu pod
cat k8s/hello-world.yaml | sed "s/\$DOCKER_IMG_VERSION/$DOCKER_IMG_VERSION/" \
    | kubectl apply -f -

kubectl get nodes -o wide -A
sleep 20
kubectl get pods -o wide -A
kubectl logs deployment/helloworld-runu |& tee /tmp/log.txt
grep "icmp_req=1" /tmp/log.txt

fold_end k8s.test.3 ""
