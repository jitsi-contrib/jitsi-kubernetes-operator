package controllers

import (
	"fmt"
	"jitsi-operator/api/v1alpha1"

	"github.com/presslabs/controller-util/syncer"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func NewProsodyServiceSyncer(jitsi *v1alpha1.Jitsi, c client.Client) syncer.Interface {
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-prosody", jitsi.Name),
			Namespace: jitsi.Namespace,
		},
	}

	return syncer.NewObjectSyncer("Service", jitsi, svc, c, func() error {
		svc.Labels = jitsi.ComponentLabels("prosody")
		svc.Spec.Type = corev1.ServiceTypeClusterIP
		svc.Spec.Selector = jitsi.ComponentLabels("prosody")
		svc.Spec.Ports = []corev1.ServicePort{

			{
				Name: "5222",
				Port: 5222,
				TargetPort: intstr.IntOrString{
					IntVal: 5222,
				},
				Protocol: corev1.ProtocolTCP,
			},
			{
				Name: "http",
				Port: 5280,
				TargetPort: intstr.IntOrString{
					IntVal: 5280,
				},
				Protocol: corev1.ProtocolTCP,
			},
			{
				Name: "external",
				Port: 5347,
				TargetPort: intstr.IntOrString{
					IntVal: 5347,
				},
				Protocol: corev1.ProtocolTCP,
			},
		}

		return nil
	})

}

func NewProsodyDeploymentSyncer(jitsi *v1alpha1.Jitsi, c client.Client) syncer.Interface {
	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-prosody", jitsi.Name),
			Namespace: jitsi.Namespace,
		},
	}

	return syncer.NewObjectSyncer("Deployment", jitsi, dep, c, func() error {
		dep.Labels = jitsi.ComponentLabels("prosody")
		dep.Spec.Template.Labels = dep.Labels
		dep.Spec.Selector = &metav1.LabelSelector{
			MatchLabels: dep.Labels,
		}
		// 	dep.Spec.Replicas = 1
		dep.Spec.Strategy.Type = appsv1.RecreateDeploymentStrategyType
		dep.Spec.Template.Spec.Containers = []corev1.Container{
			{
				Name:  "prosody",
				Image: "jitsi/prosody",
				Env: []corev1.EnvVar{
					jitsi.EnvVar("AUTH_TYPE"),
					jitsi.EnvVar("ENABLE_AUTH"),
					jitsi.EnvVar("ENABLE_GUESTS"),
					jitsi.EnvVar("ENABLE_LOBBY"),
					jitsi.EnvVar("ENABLE_XMPP_WEBSOCKET"),
					jitsi.EnvVar("GLOBAL_MODULES"),
					jitsi.EnvVar("GLOBAL_CONFIG"),
					jitsi.EnvVar("XMPP_DOMAIN"),
					jitsi.EnvVar("XMPP_AUTH_DOMAIN"),
					jitsi.EnvVar("XMPP_GUEST_DOMAIN"),
					jitsi.EnvVar("XMPP_MUC_DOMAIN"),
					jitsi.EnvVar("XMPP_INTERNAL_MUC_DOMAIN"),
					jitsi.EnvVar("XMPP_MODULES"),
					jitsi.EnvVar("XMPP_MUC_MODULES"),
					jitsi.EnvVar("XMPP_INTERNAL_MUC_MODULES"),
					jitsi.EnvVar("XMPP_RECORDER_DOMAIN"),
					jitsi.EnvVar("XMPP_CROSS_DOMAIN"),
					jitsi.EnvVar("XMPP_RECORDER_DOMAIN"),
					jitsi.EnvVar("JICOFO_COMPONENT_SECRET"),
					jitsi.EnvVar("JICOFO_AUTH_USER"),
					jitsi.EnvVar("JVB_AUTH_USER"),
					jitsi.EnvVar("JIGASI_XMPP_USER"),
					jitsi.EnvVar("JIGASI_XMPP_PASSWORD"),
					jitsi.EnvVar("JIBRI_XMPP_USER"),
					jitsi.EnvVar("JIBRI_RECORDER_USER"),
					jitsi.EnvVar("JIBRI_RECORDER_PASSWORD"),
					jitsi.EnvVar("JWT_APP_ID"),
					jitsi.EnvVar("JWT_APP_SECRET"),
					jitsi.EnvVar("JWT_ACCEPTED_ISSUERS"),
					jitsi.EnvVar("JWT_ACCEPTED_AUDIENCES"),
					jitsi.EnvVar("JWT_ASAP_KEYSERVER"),
					jitsi.EnvVar("JWT_ALLOW_EMPTY"),
					jitsi.EnvVar("JWT_AUTH_TYPE"),
					jitsi.EnvVar("JWT_TOKEN_AUTH_MODULE"),
					jitsi.EnvVar("LOG_LEVEL"),
					jitsi.EnvVar("PUBLIC_URL"),
					jitsi.EnvVar("TZ"),
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
					{
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
				},
			},
		}
		return nil
	})

}
