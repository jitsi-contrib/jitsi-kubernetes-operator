#!/bin/bash
set -e

kubectl config set-context kind-jitsi-test
cat build/jitsi-kubernetes-operator.tar | docker exec --privileged -i jitsi-test-control-plane ctr --namespace=k8s.io images import --all-platforms -
cat build/jicofo.tar | docker exec --privileged -i jitsi-test-control-plane ctr --namespace=k8s.io images import --all-platforms -
cat build/jvb.tar | docker exec --privileged -i jitsi-test-control-plane ctr --namespace=k8s.io images import --all-platforms -
cat build/prosody.tar | docker exec --privileged -i jitsi-test-control-plane ctr --namespace=k8s.io images import --all-platforms -
cat build/web.tar | docker exec --privileged -i jitsi-test-control-plane ctr --namespace=k8s.io images import --all-platforms -

kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/main/deploy/static/provider/kind/deploy.yaml

make install

LOCAL_IP=$(ip route get 1 | awk '{print $7}')
LOCAL_IP=$LOCAL_IP envsubst < ./test/jitsi.yaml | kubectl apply -f -

IMG=ghcr.io/jitsi-contrib/jitsi-kubernetes-operator:$VERSION make deploy

kubectl wait --namespace ingress-nginx \
  --for=condition=ready pod \
  --selector=app.kubernetes.io/component=controller \
  --timeout=90s

kubectl wait --namespace jitsi-operator-system \
  --for=condition=ready pod \
  --all \
  --timeout=90s

kubectl wait --namespace default \
  --for=condition=ready pod \
  --selector=app.kubernetes.io/name=jitsi,app.kubernetes.io/instance=test \
  --all \
  --timeout=90s

docker image load -i build/torture.tar
docker run --rm --add-host "test.local:$LOCAL_IP" ghcr.io/jitsi-contrib/jitsi-kubernetes-operator/torture:$VERSION -Djitsi-meet.instance.url=https://test.local -DallowInsecureCerts=true -Djitsi-meet.tests.toRun=UDPTest
