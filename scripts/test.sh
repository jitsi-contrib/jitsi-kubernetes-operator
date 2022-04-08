#!/bin/bash
set -e

kubectl config set-context kind-jitsi-test
kind load image-archive --name jitsi-test build/jitsi-kubernetes-operator.tar
kind load image-archive --name jitsi-test build/jicofo.tar
kind load image-archive --name jitsi-test build/jvb.tar
kind load image-archive --name jitsi-test build/prosody.tar
kind load image-archive --name jitsi-test build/web.tar

kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/main/deploy/static/provider/kind/deploy.yaml

cat deploy/jitsi-operator.yaml | sed "s#ghcr\.io/jitsi-contrib/jitsi-kubernetes-operator:latest#ghcr.io/jitsi-contrib/jitsi-kubernetes-operator:$VERSION#" | kubectl apply -f -

LOCAL_IP=$(ip route get 1 | awk '{print $7}')
LOCAL_IP=$LOCAL_IP envsubst < ./test/jitsi.yaml | kubectl apply -f -

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
