
# Autoscaling Kubernetes Operator

A [https://k8s.libre.sh](https://k8s.libre.sh/) project

![](https://s3.standard.indie.host/pad-liiib-re/uploads/upload_de1c8bdaf51b77c1813c18ae92ed3d7c.png)

## The issue

With covid universities and schools have to provide large scale infrastructure for visionconferences. 
From an environmental standpoint, as much as technical and fincancial, running jitsi at scale can be challenging!

## Our magic solution

Kubernetes is becoming THE cloud API, it is beautiful but hard!
This cloud infrastructure does provides the building blocks to autscale workload.

That's why we decided to pick this as the base of our solution.

Operator pattern is a way to extend easily the kubernetes API, and describe high level resources like jitsi cluster that translate into low levels resources like linux containers and network configuration.

## How it works

### Requiremenets

Depending on your region of the world, or your taste.

For the hackathon, we decided to use scaleway, as they provide autoscaling kubernetes cluster as a service

### Install our jitsi kubernetes operator
kubectl apply -f https://raw.githubusercontent.com/jitsi-contrib/jitsi-kubernetes-operator/master/deploy/jitsi-operator.yaml

### custom jitsi web interface
cf [Custom jitsi Web interface](interfaceJitsi.md)

### Profit

Now, as the critical path of a jitsi cluster are the JVBs, it will scale based on load.

### Challenges:

One JVB is deployed per node for network facilities, we need to know the JVB port
Firewall needs to allow JVB ports
A new replica of a JVB instance is a equivalent to new node in the kubernetes cluster

Single shard deployments. Multishard can be implemented later. 
1 shard = 1 signaling server - prosody and jicofo instance - and multiple JVBs and Web instances
3 Topologies:

### Static

If you wan to determistacaly define your deployements and replicas.

### Daemonset:

If all your cluster nodes are dedicated to your jitsi cluster, you can use this strategy. 
JVB processes will be deployed on each nodes. 

### Autoscalable:

JVB will be autoscaled according to stress level. 

We had to tune how to read metrics for the jvb using:

Autoscalable kubernetes cluster
Kube-metrics enabled on your cluster with [zalendo kube-metrics adapter](https://github.com/zalando-incubator/kube-metrics-adapter) provisioned
