package controllers

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/jitsi-contrib/jitsi-kubernetes-operator/api/v1alpha1"

	"github.com/presslabs/controller-util/pkg/syncer"
	"github.com/tidwall/gjson"
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
	"JICOFO_MAX_MEMORY",
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

		dep.Spec.Strategy.Type = appsv1.RecreateDeploymentStrategyType
		dep.Spec.Template.Spec.Affinity = &jitsi.Spec.Jicofo.Affinity

		envVars := append(jitsi.EnvVars(jicofoEnvs),
			corev1.EnvVar{
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
			corev1.EnvVar{
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
		)

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

		if jitsi.Spec.Metrics {
			dep.Spec.Template.Spec.Containers = append(dep.Spec.Template.Spec.Containers, NewMetricsContainer("jicofo"))
		}

		return nil
	})

}

func (r *JitsiReconciler) findJicofoPod(ctx context.Context, jitsi *v1alpha1.Jitsi) (*corev1.Pod, error) {
	pods := corev1.PodList{}
	if err := r.Client.List(ctx, &pods, client.InNamespace(jitsi.Namespace), client.MatchingLabels(jitsi.ComponentLabels("jicofo"))); err != nil {
		return nil, err
	}
	if len(pods.Items) > 0 {
		return &pods.Items[0], nil
	}
	return nil, nil
}

func (r *JitsiReconciler) getConferences(jicofo *corev1.Pod) int64 {
	if jicofo != nil && jicofo.Status.PodIP != "" {
		r.Log.Info(jicofo.Status.PodIP)
		url := fmt.Sprintf("http://%s:8888/stats", jicofo.Status.PodIP)
		res, _ := http.Get(url)
		if res != nil {
			body, err := io.ReadAll(res.Body)
			if err == nil {
				r.Log.Info(string(body))
				return gjson.Get(string(body), "conferences").Int()
			}
		}
	}
	return 0
}
