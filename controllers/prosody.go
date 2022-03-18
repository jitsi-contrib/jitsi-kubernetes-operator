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

var prosodyEnvs = []string{
	"AUTH_TYPE",
	"ENABLE_AUTH",
	"ENABLE_GUESTS",
	"ENABLE_LOBBY",
	"ENABLE_XMPP_WEBSOCKET",
	"GLOBAL_MODULES",
	"GLOBAL_CONFIG",
	"LDAP_URL",
	"LDAP_BASE",
	"LDAP_BINDDN",
	"LDAP_BINDPW",
	"LDAP_FILTER",
	"LDAP_AUTH_METHOD",
	"LDAP_VERSION",
	"LDAP_USE_TLS",
	"LDAP_TLS_CIPHERS",
	"LDAP_TLS_CHECK_PEER",
	"LDAP_TLS_CACERT_FILE",
	"LDAP_TLS_CACERT_DIR",
	"LDAP_START_TLS",
	"XMPP_DOMAIN",
	"XMPP_AUTH_DOMAIN",
	"XMPP_GUEST_DOMAIN",
	"XMPP_MUC_DOMAIN",
	"XMPP_INTERNAL_MUC_DOMAIN",
	"XMPP_MODULES",
	"XMPP_MUC_MODULES",
	"XMPP_INTERNAL_MUC_MODULES",
	"XMPP_RECORDER_DOMAIN",
	"XMPP_CROSS_DOMAIN",
	"JICOFO_AUTH_USER",
	"JVB_AUTH_USER",
	"JIGASI_XMPP_USER",
	"JIGASI_XMPP_PASSWORD",
	"JIBRI_XMPP_USER",
	"JIBRI_RECORDER_USER",
	"JWT_APP_ID",
	"JWT_APP_SECRET",
	"JWT_ACCEPTED_ISSUERS",
	"JWT_ACCEPTED_AUDIENCES",
	"JWT_ASAP_KEYSERVER",
	"JWT_ALLOW_EMPTY",
	"JWT_AUTH_TYPE",
	"JWT_TOKEN_AUTH_MODULE",
	"LOG_LEVEL",
	"PUBLIC_URL",
	"TZ",
	"ENABLE_BREAKOUT_ROOMS",
}

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

		dep.Spec.Strategy.Type = appsv1.RecreateDeploymentStrategyType
		dep.Spec.Template.Spec.Affinity = &jitsi.Spec.Prosody.Affinity

		container := corev1.Container{
			Name:            "prosody",
			Image:           jitsi.Spec.Prosody.Image,
			ImagePullPolicy: jitsi.Spec.Prosody.ImagePullPolicy,
			ReadinessProbe: &corev1.Probe{
				Handler: corev1.Handler{
					Exec: &corev1.ExecAction{
						Command: []string{
							"prosodyctl",
							"--config",
							"/config/prosody.cfg.lua",
							"status",
						},
					},
				},
			},
		}

		container.Env = append(jitsi.EnvVars(prosodyEnvs),
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

		if jitsi.Spec.TURN != nil {
			turnPreffix := "TURN"
			if jitsi.Spec.TURN.TLS {
				turnPreffix += "S"
			}

			container.Env = append(container.Env, corev1.EnvVar{
				Name:  turnPreffix + "_HOST",
				Value: jitsi.Spec.TURN.Host,
			}, corev1.EnvVar{
				Name:  turnPreffix + "_PORT",
				Value: fmt.Sprint(jitsi.Spec.TURN.Port),
			})

			if jitsi.Spec.TURN.Secret != nil {
				container.Env = append(container.Env, corev1.EnvVar{
					Name: "TURN_CREDENTIALS",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: jitsi.Spec.TURN.Secret,
					},
				})
			}
		}
		if jitsi.Spec.Prosody.CustomProsodyConfig != nil {
			dep.Spec.Template.Spec.Volumes = append(dep.Spec.Template.Spec.Volumes,
				corev1.Volume{
					Name: "jitsi-meet",
					VolumeSource: corev1.VolumeSource{
						ConfigMap: &corev1.ConfigMapVolumeSource{
							LocalObjectReference: *jitsi.Spec.Prosody.CustomProsodyConfig,
							Items: []corev1.KeyToPath{
								{
									Key:  "jitsi-meet.cfg.lua",
									Path: "jitsi-meet.cfg.lua",
								},
							},
						},
					},
				})
			container.VolumeMounts = append(container.VolumeMounts, corev1.VolumeMount{
				Name:      "jitsi-meet",
				MountPath: "/config/conf.d/jitsi-meet.cfg.lua",
				SubPath:   "jitsi-meet.cfg.lua",
			})
		}

		dep.Spec.Template.Spec.Containers = []corev1.Container{container}
		return nil
	})
}
