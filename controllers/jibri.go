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

func NewJibriDeploymentSyncer(jitsi *v1alpha1.Jitsi, c client.Client) syncer.Interface {
	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-jibri", jitsi.Name),
			Namespace: jitsi.Namespace,
		},
	}

	return syncer.NewObjectSyncer("Deployment", jitsi, dep, c, func() error {
		dep.Labels = jitsi.ComponentLabels("jibri")
		dep.Spec.Template.Labels = dep.Labels
		dep.Spec.Selector = &metav1.LabelSelector{
			MatchLabels: dep.Labels,
		}

		dep.Spec.Replicas = jitsi.Spec.Jibri.Replicas
		dep.Spec.Strategy.Type = appsv1.RecreateDeploymentStrategyType

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
		}

		jibriContainer := corev1.Container{
			Name:            "jibri",
			Image:           jitsi.Spec.Jibri.Image,
			ImagePullPolicy: jitsi.Spec.Jibri.ImagePullPolicy,
			Env: []corev1.EnvVar{
				jitsi.EnvVar("XMPP_AUTH_DOMAIN"),
				jitsi.EnvVar("XMPP_INTERNAL_MUC_DOMAIN"),
				jitsi.EnvVar("XMPP_RECORDER_DOMAIN"),
				jitsi.EnvVar("XMPP_SERVER"),
				jitsi.EnvVar("XMPP_DOMAIN"),
				jitsi.EnvVar("JIBRI_XMPP_USER"),
				jitsi.EnvVar("JIBRI_BREWERY_MUC"),
				jitsi.EnvVar("JIBRI_RECORDER_USER"),
				jitsi.EnvVar("JIBRI_FINALIZE_RECORDING_SCRIPT_PATH"),
				jitsi.EnvVar("JIBRI_STRIP_DOMAIN_JID"),
				jitsi.EnvVar("JIBRI_LOGS_DIR"),
				jitsi.EnvVar("DISPLAY"),
				jitsi.EnvVar("TZ"),
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
				{
					Name:  "DISPLAY",
					Value: "0",
				},
			},
			VolumeMounts: []corev1.VolumeMount{
				{
					Name:      "dev-snd",
					MountPath: "/dev/snd",
				},
				{
					Name:      "dev-shm",
					MountPath: "/dev/shm",
				},
			},
		}

		jibriExporterContainer := corev1.Container{
			Name:  "jibri-exporter",
			Image: "hougo13/jibri-exporter",
		}

		if jitsi.Spec.Jibri.Resources != nil {
			jibriContainer.Resources = *jitsi.Spec.Jibri.Resources
		}

		dep.Spec.Template.Spec.Containers = []corev1.Container{jibriContainer, jibriExporterContainer}

		return nil
	})

}
