package controllers

import (
	"github.com/jitsi-contrib/jitsi-kubernetes-operator/api/v1alpha1"

	"github.com/presslabs/controller-util/pkg/syncer"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

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
				Name: "recordings",
				VolumeSource: corev1.VolumeSource{
					EmptyDir: &corev1.EmptyDirVolumeSource{},
				},
			},
			{
				Name: "dev-shm",
				VolumeSource: corev1.VolumeSource{
					EmptyDir: &corev1.EmptyDirVolumeSource{
						Medium: corev1.StorageMediumMemory,
					},
				},
			},
		}

		envVars := append(jitsi.EnvVars(JibriVariables),
			corev1.EnvVar{
				Name: "JIBRI_INSTANCE_ID",
				ValueFrom: &corev1.EnvVarSource{
					FieldRef: &corev1.ObjectFieldSelector{
						FieldPath: "metadata.name",
					},
				},
			},
			corev1.EnvVar{
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
			corev1.EnvVar{
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
		)

		privileged := true
		jibriContainer := corev1.Container{
			Name:            "jibri",
			Image:           jitsi.Spec.Jibri.Image,
			ImagePullPolicy: jitsi.Spec.Jibri.ImagePullPolicy,
			Env:             envVars,
			VolumeMounts: []corev1.VolumeMount{
				{
					Name:      "recordings",
					MountPath: jitsi.EnvVarValue("JIBRI_RECORDING_DIR"),
				},
				{
					Name:      "dev-shm",
					MountPath: "/dev/shm",
				},
			},
			SecurityContext: &corev1.SecurityContext{
				Privileged: &privileged,
				Capabilities: &corev1.Capabilities{
					Add: []corev1.Capability{"NET_BIND_SERVICE", "SYS_ADMIN"},
				},
			},
		}

		if jitsi.Spec.Jibri.ServiceAccountName != "" {
			dep.Spec.Template.Spec.ServiceAccountName = jitsi.Spec.Jibri.ServiceAccountName
		}

		if jitsi.Spec.Jibri.Resources != nil {
			jibriContainer.Resources = *jitsi.Spec.Jibri.Resources
		}

		dep.Spec.Template.Spec.Containers = []corev1.Container{jibriContainer}

		if jitsi.Spec.Jibri.Bucket != nil {
			dep.Spec.Template.Spec.Containers = append(dep.Spec.Template.Spec.Containers, corev1.Container{
				Name:            "mc",
				Image:           "minio/mc:RELEASE.2022-02-23T03-15-59Z",
				ImagePullPolicy: corev1.PullIfNotPresent,
				Env: []corev1.EnvVar{
					{
						Name:  "MINIO_HOST",
						Value: jitsi.Spec.Jibri.Bucket.Host,
					},
					{
						Name:  "BUCKET",
						Value: jitsi.Spec.Jibri.Bucket.Name,
					},
					{
						Name: "MINIO_ACCESS_KEY",
						ValueFrom: &corev1.EnvVarSource{
							SecretKeyRef: &corev1.SecretKeySelector{
								Key: "ACCESS_KEY",
								LocalObjectReference: corev1.LocalObjectReference{
									Name: jitsi.Spec.Jibri.Bucket.Secret.Name,
								},
							},
						},
					},
					{
						Name: "MINIO_SECRET_KEY",
						ValueFrom: &corev1.EnvVarSource{
							SecretKeyRef: &corev1.SecretKeySelector{
								Key: "SECRET_KEY",
								LocalObjectReference: corev1.LocalObjectReference{
									Name: jitsi.Spec.Jibri.Bucket.Secret.Name,
								},
							},
						},
					},
				},
				VolumeMounts: []corev1.VolumeMount{
					{
						Name:      "recordings",
						MountPath: "/opt/recordings",
					},
				},
				Command: []string{
					"bash", "-c", "mc alias set minio $MINIO_HOST $MINIO_ACCESS_KEY $MINIO_SECRET_KEY && while true; do mc mirror --overwrite /opt/recordings minio/$BUCKET; sleep 5; done",
				},
			})
		}

		injectJibriAffinity(jitsi, &dep.Spec.Template.Spec)

		return nil
	})

}
