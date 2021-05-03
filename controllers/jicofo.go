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

func NewJicofoDeploymentSyncer(jitsi *v1alpha1.Jitsi, c client.Client) syncer.Interface {
	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-jicofo", jitsi.Name),
			Namespace: jitsi.Namespace,
		},
	}

	return syncer.NewObjectSyncer("Deployment", jitsi, dep, c, func() error {
		dep.Labels = jitsi.ComponentLabels("jicofo")
		dep.Spec.Template.Labels = dep.Labels
		dep.Spec.Selector = &metav1.LabelSelector{
			MatchLabels: dep.Labels,
		}
		// 	dep.Spec.Replicas = 1
		dep.Spec.Strategy.Type = appsv1.RecreateDeploymentStrategyType

		container := corev1.Container{
			Name:  "jicofo",
			Image: "jitsi/jicofo",
			Env: []corev1.EnvVar{
				//	jitsi.EnvVar("AUTH_TYPE"),
				//	jitsi.EnvVar("BRIDGE_AVG_PARTICIPANT_STRESS"),
				//	jitsi.EnvVar("BRIDGE_STRESS_THRESHOLD"),
				//	jitsi.EnvVar("ENABLE_AUTH"),
				//	jitsi.EnvVar("ENABLE_AUTO_OWNER"),
				//	jitsi.EnvVar("ENABLE_CODEC_VP8"),
				//	jitsi.EnvVar("ENABLE_CODEC_VP9"),
				//	jitsi.EnvVar("ENABLE_CODEC_H264"),
				jitsi.EnvVar("ENABLE_OCTO"),
				jitsi.EnvVar("ENABLE_RECORDING"),
				//	jitsi.EnvVar("ENABLE_SCTP"),
				jitsi.EnvVar("JICOFO_AUTH_USER"),
				//	jitsi.EnvVar("JICOFO_AUTH_PASSWORD"),
				//	jitsi.EnvVar("JICOFO_ENABLE_BRIDGE_HEALTH_CHECKS"),
				//	jitsi.EnvVar("JICOFO_CONF_INITIAL_PARTICIPANT_WAIT_TIMEOUT"),
				//	jitsi.EnvVar("JICOFO_CONF_SINGLE_PARTICIPANT_TIMEOUT"),
				//	jitsi.EnvVar("JICOFO_ENABLE_HEALTH_CHECKS"),
				//	jitsi.EnvVar("JICOFO_SHORT_ID"),
				//	jitsi.EnvVar("JICOFO_RESERVATION_ENABLED"),
				//	jitsi.EnvVar("JICOFO_RESERVATION_REST_BASE_URL"),
				jitsi.EnvVar("JIBRI_BREWERY_MUC"),
				//		jitsi.EnvVar("JIBRI_REQUEST_RETRIES"),
				jitsi.EnvVar("JIBRI_PENDING_TIMEOUT"),
				//		jitsi.EnvVar("JIGASI_BREWERY_MUC"),
				//		jitsi.EnvVar("JIGASI_SIP_URI"),
				jitsi.EnvVar("JVB_BREWERY_MUC"),
				//	jitsi.EnvVar("MAX_BRIDGE_PARTICIPANTS"),
				jitsi.EnvVar("OCTO_BRIDGE_SELECTION_STRATEGY"),
				jitsi.EnvVar("TZ"),
				jitsi.EnvVar("XMPP_DOMAIN"),
				jitsi.EnvVar("XMPP_AUTH_DOMAIN"),
				jitsi.EnvVar("XMPP_INTERNAL_MUC_DOMAIN"),
				jitsi.EnvVar("XMPP_MUC_DOMAIN"),
				jitsi.EnvVar("XMPP_SERVER"),
				{
					Name: "JICOFO_COMPONENT_SECRET",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: jitsi.Name,
							},
							Key: "JICOFO_COMPONENT_SECRET",
						},
					},
				},
				{
					Name: "JICOFO_AUTH_PASSWORD",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: jitsi.Name,
							},
							Key: "JICOFO_AUTH_PASSWORD",
						},
					},
				},
			},
		}

		if jitsi.Spec.Jicofo.Resources != nil {
			container.Resources = *jitsi.Spec.Jicofo.Resources
		}

		dep.Spec.Template.Spec.Containers = []corev1.Container{container}
		return nil
	})

}
