package controllers

import (
	"fmt"
	"jitsi-operator/api/v1alpha1"

	"github.com/presslabs/controller-util/rand"
	"github.com/presslabs/controller-util/syncer"
	autoscalingv2 "k8s.io/api/autoscaling/v2beta2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
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
org.jitsi.videobridge.ENABLE_REST_SHUTDOWN=true
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
				Name:     "udp",
				Port:     *jitsi.Spec.JVB.Ports.UDP,
				NodePort: *jitsi.Spec.JVB.Ports.UDP,
				TargetPort: intstr.IntOrString{
					IntVal: *jitsi.Spec.JVB.Ports.UDP,
				},
				Protocol: corev1.ProtocolUDP,
			},
			{
				Name:     "tcp",
				Port:     *jitsi.Spec.JVB.Ports.TCP,
				NodePort: *jitsi.Spec.JVB.Ports.TCP,
				TargetPort: intstr.IntOrString{
					IntVal: *jitsi.Spec.JVB.Ports.TCP,
				},
				Protocol: corev1.ProtocolTCP,
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

func injectJVBAffinity(jitsi *v1alpha1.Jitsi, pod *corev1.PodSpec) {
	if jitsi.Spec.JVB.DisableDefaultAffinity {
		pod.Affinity = &jitsi.Spec.Jibri.Affinity
	} else {
		pod.Affinity = &corev1.Affinity{
			PodAntiAffinity: &corev1.PodAntiAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
					{
						LabelSelector: &metav1.LabelSelector{
							MatchLabels: jitsi.ComponentLabels("jvb"),
						},
						TopologyKey: "kubernetes.io/hostname",
					},
				},
				PreferredDuringSchedulingIgnoredDuringExecution: []corev1.WeightedPodAffinityTerm{
					{
						Weight: 100,
						PodAffinityTerm: corev1.PodAffinityTerm{
							LabelSelector: &metav1.LabelSelector{
								MatchLabels: jitsi.ComponentLabels("jibri"),
							},
							TopologyKey: "kubernetes.io/hostname",
						},
					},
				},
			},
		}
		MergeAffinities(pod.Affinity, jitsi.Spec.JVB.Affinity)
	}

}

func NewJVBDeploymentSyncer(jitsi *v1alpha1.Jitsi, c client.Client) syncer.Interface {
	dep := jitsi.JVBDeployment()

	return syncer.NewObjectSyncer("Deployment", jitsi, &dep, c, func() error {
		dep.Labels = jitsi.ComponentLabels("jvb")

		jitsi.JVBPodTemplateSpec(&dep.Spec.Template)

		dep.Spec.Template.Labels = dep.Labels

		dep.Spec.Selector = &metav1.LabelSelector{
			MatchLabels: dep.Labels,
		}

		injectJVBAffinity(jitsi, &dep.Spec.Template.Spec)

		dep.Spec.Replicas = jitsi.Spec.JVB.Strategy.Replicas

		return nil
	})

}

func NewJVBHPASyncer(jitsi *v1alpha1.Jitsi, c client.Client) syncer.Interface {
	obj := jitsi.JVBHPA()

	return syncer.NewObjectSyncer("HorizontalPodAutoscaler", jitsi, &obj, c, func() error {
		obj.Labels = jitsi.ComponentLabels("jvb")

		obj.Annotations = make(map[string]string)
		obj.Annotations["metric-config.pods.jitsi-stress-level.json-path/json-key"] = "$.stress_level"
		obj.Annotations["metric-config.pods.jitsi-stress-level.json-path/path"] = "/colibri/stats"
		obj.Annotations["metric-config.pods.jitsi-stress-level.json-path/port"] = "8080"

		obj.Spec = autoscalingv2.HorizontalPodAutoscalerSpec{
			ScaleTargetRef: autoscalingv2.CrossVersionObjectReference{
				APIVersion: "apps/v1",
				Kind:       "Deployment",
				Name:       fmt.Sprintf("%s-jvb", jitsi.Name),
			},
			MinReplicas: jitsi.Spec.JVB.Strategy.Replicas,
			MaxReplicas: jitsi.Spec.JVB.Strategy.MaxReplicas,
			Metrics: []autoscalingv2.MetricSpec{
				{
					Type: autoscalingv2.PodsMetricSourceType,
					Pods: &autoscalingv2.PodsMetricSource{
						Metric: autoscalingv2.MetricIdentifier{
							Name: "jitsi-stress-level",
						},
						Target: autoscalingv2.MetricTarget{
							Type:         autoscalingv2.AverageValueMetricType,
							AverageValue: resource.NewMilliQuantity(10, resource.DecimalSI),
						},
					},
				},
			},
		}

		return nil
	})

}

func NewJVBDaemonSetSyncer(jitsi *v1alpha1.Jitsi, c client.Client) syncer.Interface {
	dep := jitsi.JVBDaemonSet()

	return syncer.NewObjectSyncer("DaemonSet", jitsi, &dep, c, func() error {
		dep.Labels = jitsi.ComponentLabels("jvb")

		jitsi.JVBPodTemplateSpec(&dep.Spec.Template)

		dep.Spec.Template.Labels = dep.Labels

		dep.Spec.Selector = &metav1.LabelSelector{
			MatchLabels: dep.Labels,
		}

		injectJVBAffinity(jitsi, &dep.Spec.Template.Spec)

		return nil
	})

}
