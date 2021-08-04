package controllers

import (
	"jitsi-operator/api/v1alpha1"

	"github.com/presslabs/controller-util/syncer"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var jibriEnvs = []string{
	"PUBLIC_URL",
	"XMPP_AUTH_DOMAIN",
	"XMPP_INTERNAL_MUC_DOMAIN",
	"XMPP_RECORDER_DOMAIN",
	"XMPP_SERVER",
	"XMPP_DOMAIN",
	"JIBRI_XMPP_USER",
	"JIBRI_BREWERY_MUC",
	"JIBRI_RECORDER_USER",
	"JIBRI_RECORDING_DIR",
	"JIBRI_FINALIZE_RECORDING_SCRIPT_PATH",
	"JIBRI_STRIP_DOMAIN_JID",
	"JIBRI_LOGS_DIR",
	"DISPLAY",
	"TZ",
}

func injectJibriAffinity(jitsi *v1alpha1.Jitsi, pod *corev1.PodSpec) {
	if jitsi.Spec.Jibri.DisableDefaultAffinity {
		pod.Affinity = &jitsi.Spec.Jibri.Affinity
	} else {
		pod.Affinity = &corev1.Affinity{
			PodAntiAffinity: &corev1.PodAntiAffinity{
				PreferredDuringSchedulingIgnoredDuringExecution: []corev1.WeightedPodAffinityTerm{
					{
						Weight: 100,
						PodAffinityTerm: corev1.PodAffinityTerm{
							LabelSelector: &metav1.LabelSelector{
								MatchLabels: jitsi.ComponentLabels("jvb"),
							},
							TopologyKey: "kubernetes.io/hostname",
						},
					},
				},
			},
		}
		MergeAffinities(pod.Affinity, jitsi.Spec.Jibri.Affinity)
	}

}

func NewJibriDeploymentSyncer(jitsi *v1alpha1.Jitsi, c client.Client) syncer.Interface {
	dep := jitsi.JibriDeployment()

	return syncer.NewObjectSyncer("Deployment", jitsi, &dep, c, func() error {
		dep.Labels = jitsi.ComponentLabels("jibri")
		dep.Spec.Template.Labels = dep.Labels
		dep.Spec.Selector = &metav1.LabelSelector{
			MatchLabels: dep.Labels,
		}

		dep.Spec.Replicas = jitsi.Spec.Jibri.Replicas
		dep.Spec.Strategy.Type = appsv1.RollingUpdateDeploymentStrategyType

		dep.Spec.Template.Spec.Volumes = []corev1.Volume{
			{
				Name: "dev-snd",
				VolumeSource: corev1.VolumeSource{
					HostPath: &corev1.HostPathVolumeSource{
						Path: "/dev/snd",
					},
				},
			},
			{
				Name: "dev-shm",
				VolumeSource: corev1.VolumeSource{
					HostPath: &corev1.HostPathVolumeSource{
						Path: "/dev/shm",
					},
				},
			},
			{
				Name: "recordings",
				VolumeSource: corev1.VolumeSource{
					EmptyDir: &corev1.EmptyDirVolumeSource{},
				},
			},
		}

		envVars := []corev1.EnvVar{
			{
				Name: "JIBRI_INSTANCE_ID",
				ValueFrom: &corev1.EnvVarSource{
					FieldRef: &corev1.ObjectFieldSelector{
						FieldPath: "metadata.name",
					},
				},
			},
			{
				Name: "JIBRI_XMPP_PASSWORD",
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: jitsi.Name,
						},
						Key: "JIBRI_XMPP_PASSWORD",
					},
				},
			},
			{
				Name: "JIBRI_RECORDER_PASSWORD",
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: jitsi.Name,
						},
						Key: "JIBRI_RECORDER_PASSWORD",
					},
				},
			},
		}

		for _, env := range jibriEnvs {
			if len(jitsi.EnvVar(env).Value) > 0 {
				envVars = append(envVars, jitsi.EnvVar(env))
			}
		}

		privileged := true
		jibriContainer := corev1.Container{
			Name:            "jibri",
			Image:           jitsi.Spec.Jibri.Image,
			ImagePullPolicy: jitsi.Spec.Jibri.ImagePullPolicy,
			Env:             envVars,
			VolumeMounts: []corev1.VolumeMount{
				{
					Name:      "dev-snd",
					MountPath: "/dev/snd",
				},
				{
					Name:      "dev-shm",
					MountPath: "/dev/shm",
				},
				{
					Name:      "recordings",
					MountPath: jitsi.EnvVarValue("JIBRI_RECORDING_DIR"),
				},
			},
			SecurityContext: &corev1.SecurityContext{
				Privileged: &privileged,
				Capabilities: &corev1.Capabilities{
					Add: []corev1.Capability{"NET_BIND_SERVICE", "SYS_ADMIN"},
				},
			},
		}

		if jitsi.Spec.Jibri.Resources != nil {
			jibriContainer.Resources = *jitsi.Spec.Jibri.Resources
		}

		dep.Spec.Template.Spec.Containers = []corev1.Container{jibriContainer}

		injectJibriAffinity(jitsi, &dep.Spec.Template.Spec)

		return nil
	})

}
