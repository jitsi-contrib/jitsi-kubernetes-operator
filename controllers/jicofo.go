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

var jicofoEnvs = []string{
	"AUTH_TYPE",
	"BRIDGE_AVG_PARTICIPANT_STRESS",
	"BRIDGE_STRESS_THRESHOLD",
	"ENABLE_AUTH",
	"ENABLE_AUTO_OWNER",
	"ENABLE_CODEC_VP8",
	"ENABLE_CODEC_VP9",
	"ENABLE_CODEC_H264",
	"ENABLE_OCTO",
	"ENABLE_RECORDING",
	"ENABLE_SCTP",
	"JICOFO_AUTH_USER",
	"JICOFO_AUTH_PASSWORD",
	"JICOFO_ENABLE_BRIDGE_HEALTH_CHECKS",
	"JICOFO_CONF_INITIAL_PARTICIPANT_WAIT_TIMEOUT",
	"JICOFO_CONF_SINGLE_PARTICIPANT_TIMEOUT",
	"JICOFO_ENABLE_HEALTH_CHECKS",
	"JICOFO_SHORT_ID",
	"JICOFO_RESERVATION_ENABLED",
	"JICOFO_RESERVATION_REST_BASE_URL",
	"JIBRI_BREWERY_MUC",
	"JIBRI_REQUEST_RETRIES",
	"JIBRI_PENDING_TIMEOUT",
	"JIGASI_BREWERY_MUC",
	"JIGASI_SIP_URI",
	"JVB_BREWERY_MUC",
	"MAX_BRIDGE_PARTICIPANTS",
	"OCTO_BRIDGE_SELECTION_STRATEGY",
	"TZ",
	"XMPP_DOMAIN",
	"XMPP_AUTH_DOMAIN",
	"XMPP_INTERNAL_MUC_DOMAIN",
	"XMPP_MUC_DOMAIN",
	"XMPP_SERVER",
}

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
		dep.Spec.Template.Spec.Affinity = &jitsi.Spec.Jicofo.Affinity

		envVars := []corev1.EnvVar{
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
		}

		for _, env := range jicofoEnvs {
			if len(jitsi.EnvVar(env).Value) > 0 {
				envVars = append(envVars, jitsi.EnvVar(env))
			}
		}

		container := corev1.Container{
			Name:            "jicofo",
			Image:           jitsi.Spec.Jicofo.Image,
			ImagePullPolicy: jitsi.Spec.Jicofo.ImagePullPolicy,
			Env:             envVars,
		}

		if jitsi.Spec.Jicofo.Resources != nil {
			container.Resources = *jitsi.Spec.Jicofo.Resources
		}

		dep.Spec.Template.Spec.Containers = []corev1.Container{container}
		return nil
	})

}
