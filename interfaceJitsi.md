# customize jitsi web interface

## Requirements:

- install go 1.15.5
- install controller-gen
- install gcc and build-base

- How to build your own manager (custom jitsi web interface)
    * go to Makefile: change IMG ?= with your own repository docker 
    * make build
    * make docker-build
    * make docker-push 
    * make generate-deploy
    * make deploy
    * kubectl apply -f deploy/jitsi-operator
- Install your configmap and your Jitsi stack (cf config/samples/mydomain.com_jitsi_instance.yaml  config/samples/mydomain.com_jitsi_instance.yaml)