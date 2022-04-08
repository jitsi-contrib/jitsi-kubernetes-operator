package controllers

import (
	"fmt"
	"jitsi-operator/api/v1alpha1"

	"github.com/presslabs/controller-util/pkg/rand"
	"github.com/presslabs/controller-util/pkg/syncer"
	autoscalingv2 "k8s.io/api/autoscaling/v2beta2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const jvbConf = `
org.ice4j.ice.harvest.DISABLE_AWS_HARVESTER=true
`

var secretsVar = []string{
	"JICOFO_COMPONENT_SECRET",
	"JICOFO_AUTH_PASSWORD",
	"JVB_AUTH_PASSWORD",
	"JIBRI_XMPP_PASSWORD",
	"JIBRI_RECORDER_PASSWORD",
}

var jvbEnvs = []string{
	"ENABLE_COLIBRI_WEBSOCKET",
	"ENABLE_OCTO",
	"DOCKER_HOST_ADDRESS",
	"XMPP_AUTH_DOMAIN",
	"XMPP_INTERNAL_MUC_DOMAIN",
	"XMPP_SERVER",
	"JVB_AUTH_USER",
	"JVB_BREWERY_MUC",
	"JVB_PORT",
	"JVB_TCP_HARVESTER_DISABLED",
	"JVB_TCP_PORT",
	"JVB_TCP_MAPPED_PORT",
	"JVB_STUN_SERVERS",
	"COLIBRI_REST_ENABLED",
	"JVB_WS_DOMAIN",
	"JVB_WS_SERVER_ID",
	"PUBLIC_URL",
	"JVB_OCTO_BIND_ADDRESS",
	"JVB_OCTO_PUBLIC_ADDRESS",
	"JVB_OCTO_BIND_PORT",
	"JVB_OCTO_REGION",
	"TZ",
	"SHUTDOWN_REST_ENABLED",
	"VIDEOBRIDGE_MAX_MEMORY",
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
		pod.Affinity = &jitsi.Spec.JVB.Affinity
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

func JVBPodTemplateSpec(jitsi *v1alpha1.Jitsi, podSpec *corev1.PodTemplateSpec) {
	podSpec.Spec.Volumes = []corev1.Volume{
		{
			Name: "jvb-config",
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: fmt.Sprintf("%s-jvb", jitsi.Name),
					},
				},
			},
		},
		{
			Name: "config",
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		},
	}

	podSpec.Spec.InitContainers = []corev1.Container{
		{
			Name:  "config",
			Image: "busybox:stable",
			Command: []string{
				"cp", "-f", "/config-src/sip-communicator.properties", "/config/custom-sip-communicator.properties",
			},
			VolumeMounts: []corev1.VolumeMount{
				{
					Name:      "jvb-config",
					MountPath: "/config-src",
				},
				{
					Name:      "config",
					MountPath: "/config",
				},
			},
		},
	}

	envVars := append(jitsi.EnvVars(jvbEnvs),
		corev1.EnvVar{
			Name: "LOCAL_ADDRESS",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					FieldPath: "status.podIP",
				},
			},
		},
		corev1.EnvVar{
			Name: "JVB_AUTH_PASSWORD",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: jitsi.Name,
					},
					Key: "JVB_AUTH_PASSWORD",
				},
			},
		},
		// TODO options to manage host IP through vars or stun servers
		// {
		// 	Name: "DOCKER_HOST_ADDRESS",
		// 	ValueFrom: &corev1.EnvVarSource{
		// 		FieldRef: &corev1.ObjectFieldSelector{
		// 			FieldPath: "status.hostIP",
		// 		},
		// 	},
		// },

		// {
		// Default 0.0.0.0
		// 	Name: "JVB_OCTO_BIND_ADDRESS",
		// 	ValueFrom: &corev1.EnvVarSource{
		// 		FieldRef: &corev1.ObjectFieldSelector{
		// 			FieldPath: "status.podIP",
		// 		},
		// 	},
		// },

		corev1.EnvVar{
			Name: "JVB_OCTO_PUBLIC_ADDRESS",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					FieldPath: "status.podIP",
				},
			},
		},
		corev1.EnvVar{
			Name: "JVB_WS_SERVER_ID",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					FieldPath: "status.podIP",
				},
			},
		},
	)

	jvbContainer := corev1.Container{
		Name:            "jvb",
		Image:           jitsi.Spec.JVB.Image,
		ImagePullPolicy: jitsi.Spec.JVB.ImagePullPolicy,
		Env:             envVars,
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      "config",
				MountPath: "/config",
			},
		},
		Ports: []corev1.ContainerPort{
			{
				Name:          "rtp-udp",
				ContainerPort: *jitsi.Spec.JVB.Ports.UDP,
				HostPort:      *jitsi.Spec.JVB.Ports.UDP,
				Protocol:      corev1.ProtocolUDP,
			},
		},
		ReadinessProbe: &corev1.Probe{
			ProbeHandler: corev1.ProbeHandler{
				HTTPGet: &corev1.HTTPGetAction{
					Path: "/about/health",
					Port: intstr.FromInt(8080),
				},
			},
			InitialDelaySeconds: 10,
		},
	}

	if jitsi.Spec.JVB.GracefulShutdown {
		jvbContainer.Lifecycle = &corev1.Lifecycle{
			PreStop: &corev1.LifecycleHandler{
				Exec: &corev1.ExecAction{
					Command: []string{
						"bash", "-c", "/usr/share/jitsi-videobridge/graceful_shutdown.sh -p $(s6-svstat -o pid /var/run/s6/services/jvb) -t 3 -s",
					},
				},
			},
		}
	}

	if jitsi.Spec.JVB.Resources != nil {
		jvbContainer.Resources = *jitsi.Spec.JVB.Resources
	}

	podSpec.Spec.Containers = []corev1.Container{jvbContainer}

	if jitsi.Spec.Metrics {
		podSpec.Spec.Containers = append(podSpec.Spec.Containers, NewMetricsContainer("jvb"))
	}
}

func NewJVBDeploymentSyncer(jitsi *v1alpha1.Jitsi, c client.Client) syncer.Interface {
	dep := jitsi.JVBDeployment()

	return syncer.NewObjectSyncer("Deployment", jitsi, &dep, c, func() error {
		dep.Labels = jitsi.ComponentLabels("jvb")

		JVBPodTemplateSpec(jitsi, &dep.Spec.Template)

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

		JVBPodTemplateSpec(jitsi, &dep.Spec.Template)

		dep.Spec.Template.Labels = dep.Labels

		dep.Spec.Selector = &metav1.LabelSelector{
			MatchLabels: dep.Labels,
		}

		injectJVBAffinity(jitsi, &dep.Spec.Template.Spec)

		return nil
	})

}
