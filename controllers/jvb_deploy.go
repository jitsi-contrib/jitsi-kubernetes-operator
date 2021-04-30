package controllers

import (
	"fmt"
	"jitsi-operator/api/v1alpha1"

	"github.com/presslabs/controller-util/syncer"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const jvbConf = `
{{ if .Env.DOCKER_HOST_ADDRESS }}
org.ice4j.ice.harvest.NAT_HARVESTER_LOCAL_ADDRESS={{ .Env.LOCAL_ADDRESS }}
org.ice4j.ice.harvest.NAT_HARVESTER_PUBLIC_ADDRESS={{ .Env.DOCKER_HOST_ADDRESS }}
{{ end }}
org.ice4j.ice.harvest.DISABLE_AWS_HARVESTER=true

org.jitsi.videobridge.octo.BIND_ADDRESS={{ .Env.LOCAL_ADDRESS }}
org.jitsi.videobridge.octo.BIND_PORT=4096
org.jitsi.videobridge.REGION={{ .Env.DEPLOYMENTINFO_USERREGION }}
`

func NewJVBConfigMapSyncer(jitsi *v1alpha1.Jitsi, c client.Client) syncer.Interface {

	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-jvb", jitsi.Name),
			Namespace: jitsi.Namespace,
		},
	}

	return syncer.NewObjectSyncer("ConfigMap", jitsi, cm, c, func() error {
		cm.Labels = jitsi.ComponentLabels("jvb")
		cm.Data = map[string]string{
			"sip-communicator.properties": jvbConf,
		}

		return nil
	})

}

func NewJVBDeploymentSyncer(jitsi *v1alpha1.Jitsi, c client.Client) syncer.Interface {

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-jvb", jitsi.Name),
			Namespace: jitsi.Namespace,
		},
	}

	return syncer.NewObjectSyncer("Deployment", jitsi, dep, c, func() error {
		dep.Labels = jitsi.ComponentLabels("jvb")

		jitsi.JVBPodTemplateSpec(&dep.Spec.Template)

		dep.Spec.Template.Labels = dep.Labels

		dep.Spec.Selector = &metav1.LabelSelector{
			MatchLabels: dep.Labels,
		}

		dep.Spec.Replicas = &jitsi.Spec.JVB.Strategy.Replicas

		return nil
	})

}
