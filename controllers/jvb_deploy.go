package controllers

import (
	"fmt"
	"jitsi-operator/api/v1alpha1"

	"github.com/presslabs/controller-util/rand"
	"github.com/presslabs/controller-util/syncer"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
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

var secretsVar = []string{
	"JICOFO_COMPONENT_SECRET",
	"JICOFO_AUTH_PASSWORD",
	"JVB_AUTH_PASSWORD",
	"JIBRI_XMPP_PASSWORD",
	"JIBRI_RECORDER_PASSWORD",
}

func NewJitsiSecretSyncer(jitsi *v1alpha1.Jitsi, c client.Client) syncer.Interface {
	sec := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      jitsi.Name,
			Namespace: jitsi.Namespace,
		},
	}

	return syncer.NewObjectSyncer("Secret", jitsi, sec, c, func() error {
		sec.Labels = jitsi.ComponentLabels("core")

		if len(sec.Data) == 0 {
			sec.Data = make(map[string][]byte, 5)
		}

		for _, secretVar := range secretsVar {
			if len(sec.Data[secretVar]) == 0 {
				random, err := rand.AlphaNumericString(32)
				if err != nil {
					return err
				}

				sec.Data[secretVar] = []byte(random)
			}
		}

		return nil
	})

}

func NewJVBServiceSyncer(jitsi *v1alpha1.Jitsi, c client.Client) syncer.Interface {
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-jvb", jitsi.Name),
			Namespace: jitsi.Namespace,
		},
	}

	return syncer.NewObjectSyncer("Service", jitsi, svc, c, func() error {
		svc.Labels = jitsi.ComponentLabels("jvb")
		svc.Spec.Type = corev1.ServiceTypeNodePort
		svc.Spec.Ports = []corev1.ServicePort{
			{
				Name: "udp",
				Port: *jitsi.Spec.JVB.Ports.UDP,
				TargetPort: intstr.IntOrString{
					IntVal: *jitsi.Spec.JVB.Ports.UDP,
				},
				Protocol: corev1.ProtocolUDP,
			},
		}

		svc.Spec.Selector = jitsi.ComponentLabels("jvb")

		return nil
	})

}

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

		dep.Spec.Replicas = jitsi.Spec.JVB.Strategy.Replicas

		return nil
	})

}
