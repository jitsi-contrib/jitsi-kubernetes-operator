#!/bin/bash
set -e

JITSI_VERSION=$(cat JITSI_VERSION)

for context in images/*; do
    name=$(basename $context)
    docker buildx build --build-arg "JITSI_VERSION=$JITSI_VERSION" --file images/$name/Containerfile -t ghcr.io/jitsi-contrib/jitsi-kubernetes-operator/$name:$VERSION -o type=docker,dest=build/$name.tar $context
done

docker buildx build --build-arg "VERSION=$VERSION" -t ghcr.io/jitsi-contrib/jitsi-kubernetes-operator:$VERSION -o type=docker,dest=build/jitsi-kubernetes-operator.tar .