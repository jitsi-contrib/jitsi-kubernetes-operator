package controllers

import (
	"fmt"

	"github.com/jitsi-contrib/jitsi-kubernetes-operator/api/v1alpha1"

	"github.com/presslabs/controller-util/pkg/syncer"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var webEnvs = []string{
	"AMPLITUDE_ID",
	"ANALYTICS_SCRIPT_URLS",
	"ANALYTICS_WHITELISTED_EVENTS",
	"AUDIO_QUALITY_OPUS_BITRATE",
	"BRANDING_DATA_URL",
	"CALLSTATS_CUSTOM_SCRIPT_URL",
	"CALLSTATS_ID",
	"CHROME_EXTENSION_BANNER_JSON",
	"CONFCODE_URL",
	"CONFIG_EXTERNAL_CONNECT",
	"DEFAULT_LANGUAGE",
	"DEPLOYMENTINFO_ENVIRONMENT",
	"DEPLOYMENTINFO_ENVIRONMENT_TYPE",
	"DEPLOYMENTINFO_REGION",
	"DEPLOYMENTINFO_SHARD",
	"DEPLOYMENTINFO_USERREGION",
	"DESKTOP_SHARING_FRAMERATE_MIN",
	"DESKTOP_SHARING_FRAMERATE_MAX",
	"DIALIN_NUMBERS_URL",
	"DIALOUT_AUTH_URL",
	"DIALOUT_CODES_URL",
	"DISABLE_AUDIO_LEVELS",
	"DISABLE_DEEP_LINKING",
	"DISABLE_GRANT_MODERATOR",
	"DISABLE_HTTPS",
	"DISABLE_KICKOUT",
	"DISABLE_LOCAL_RECORDING",
	"DISABLE_POLLS",
	"DISABLE_PRIVATE_CHAT",
	"DISABLE_PROFILE",
	"DISABLE_REACTIONS",
	"DISABLE_REMOTE_VIDEO_MENU",
	"DROPBOX_APPKEY",
	"DROPBOX_REDIRECT_URI",
	"DYNAMIC_BRANDING_URL",
	"ENABLE_AUDIO_PROCESSING",
	"ENABLE_AUTH",
	"ENABLE_BREAKOUT_ROOMS",
	"ENABLE_CALENDAR",
	"ENABLE_COLIBRI_WEBSOCKET",
	"ENABLE_E2EPING",
	"ENABLE_FILE_RECORDING_SHARING",
	"ENABLE_GUESTS",
	"ENABLE_HSTS",
	"ENABLE_HTTP_REDIRECT",
	"ENABLE_IPV6",
	"ENABLE_LETSENCRYPT",
	"ENABLE_LIPSYNC",
	"ENABLE_NO_AUDIO_DETECTION",
	"ENABLE_NOISY_MIC_DETECTION",
	"ENABLE_OPUS_RED",
	"ENABLE_PREJOIN_PAGE",
	"ENABLE_P2P",
	"ENABLE_WELCOME_PAGE",
	"ENABLE_CLOSE_PAGE",
	"ENABLE_LOCAL_RECORDING_NOTIFY_ALL_PARTICIPANT",
	"ENABLE_LOCAL_RECORDING_SELF_START",
	"ENABLE_RECORDING",
	"ENABLE_REMB",
	"ENABLE_REQUIRE_DISPLAY_NAME",
	"ENABLE_SIMULCAST",
	"ENABLE_STATS_ID",
	"ENABLE_STEREO",
	"ENABLE_SUBDOMAINS",
	"ENABLE_TALK_WHILE_MUTED",
	"ENABLE_TCC",
	"ENABLE_TRANSCRIPTIONS",
	"ENABLE_XMPP_WEBSOCKET",
	"ENABLE_JAAS_COMPONENTS",
	"ENABLE_MULTI_STREAM",
	"ETHERPAD_PUBLIC_URL",
	"ETHERPAD_URL_BASE",
	"E2EPING_NUM_REQUESTS",
	"E2EPING_MAX_CONFERENCE_SIZE",
	"E2EPING_MAX_MESSAGE_PER_SECOND",
	"GOOGLE_ANALYTICS_ID",
	"GOOGLE_API_APP_CLIENT_ID",
	"HIDE_PREMEETING_BUTTONS",
	"HIDE_PREJOIN_DISPLAY_NAME",
	"HIDE_PREJOIN_EXTRA_BUTTONS",
	"INVITE_SERVICE_URL",
	"JICOFO_AUTH_USER",
	"LETSENCRYPT_DOMAIN",
	"LETSENCRYPT_EMAIL",
	"LETSENCRYPT_USE_STAGING",
	"MATOMO_ENDPOINT",
	"MATOMO_SITE_ID",
	"MICROSOFT_API_APP_CLIENT_ID",
	"NGINX_RESOLVER",
	"NGINX_WORKER_PROCESSES",
	"NGINX_WORKER_CONNECTIONS",
	"PEOPLE_SEARCH_URL",
	"PUBLIC_URL",
	"P2P_PREFERRED_CODEC",
	"RESOLUTION",
	"RESOLUTION_MIN",
	"RESOLUTION_WIDTH",
	"RESOLUTION_WIDTH_MIN",
	"START_AUDIO_MUTED",
	"START_AUDIO_ONLY",
	"START_BITRATE",
	"START_SILENT",
	"START_WITH_AUDIO_MUTED",
	"START_VIDEO_MUTED",
	"START_WITH_VIDEO_MUTED",
	"TESTING_CAP_SCREENSHARE_BITRATE",
	"TESTING_OCTO_PROBABILITY",
	"TOKEN_AUTH_URL",
	"TOOLBAR_BUTTONS",
	"TZ",
	"VIDEOQUALITY_BITRATE_H264_LOW",
	"VIDEOQUALITY_BITRATE_H264_STANDARD",
	"VIDEOQUALITY_BITRATE_H264_HIGH",
	"VIDEOQUALITY_BITRATE_VP8_LOW",
	"VIDEOQUALITY_BITRATE_VP8_STANDARD",
	"VIDEOQUALITY_BITRATE_VP8_HIGH",
	"VIDEOQUALITY_BITRATE_VP9_LOW",
	"VIDEOQUALITY_BITRATE_VP9_STANDARD",
	"VIDEOQUALITY_BITRATE_VP9_HIGH",
	"VIDEOQUALITY_ENFORCE_PREFERRED_CODEC",
	"VIDEOQUALITY_PREFERRED_CODEC",
	"XMPP_AUTH_DOMAIN",
	"XMPP_BOSH_URL_BASE",
	"XMPP_DOMAIN",
	"XMPP_GUEST_DOMAIN",
	"XMPP_MUC_DOMAIN",
	"XMPP_RECORDER_DOMAIN",
	"XMPP_PORT",
}

func NewWebServiceSyncer(jitsi *v1alpha1.Jitsi, c client.Client) syncer.Interface {
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-web", jitsi.Name),
			Namespace: jitsi.Namespace,
		},
	}

	return syncer.NewObjectSyncer("Service", jitsi, svc, c, func() error {
		svc.Labels = jitsi.ComponentLabels("web")
		svc.Spec.Type = corev1.ServiceTypeClusterIP
		svc.Spec.Selector = jitsi.ComponentLabels("web")
		svc.Spec.Ports = []corev1.ServicePort{
			{
				Name: "http",
				Port: 80,
				TargetPort: intstr.IntOrString{
					IntVal: 80,
				},
				Protocol: corev1.ProtocolTCP,
			},
			{
				Name: "https",
				Port: 443,
				TargetPort: intstr.IntOrString{
					IntVal: 443,
				},
				Protocol: corev1.ProtocolTCP,
			},
		}

		return nil
	})

}

func NewWebDeploymentSyncer(jitsi *v1alpha1.Jitsi, c client.Client) syncer.Interface {
	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-web", jitsi.Name),
			Namespace: jitsi.Namespace,
		},
	}

	return syncer.NewObjectSyncer("Deployment", jitsi, dep, c, func() error {
		dep.Labels = jitsi.ComponentLabels("web")
		dep.Spec.Template.Labels = dep.Labels
		dep.Spec.Selector = &metav1.LabelSelector{
			MatchLabels: dep.Labels,
		}

		dep.Spec.Replicas = jitsi.Spec.Web.Replicas
		dep.Spec.Strategy.Type = appsv1.RollingUpdateDeploymentStrategyType
		dep.Spec.Template.Spec.Affinity = &jitsi.Spec.Web.Affinity

		envVars := jitsi.EnvVars(webEnvs)

		container := corev1.Container{
			Name:            "web",
			Image:           jitsi.Spec.Web.Image,
			ImagePullPolicy: jitsi.Spec.Web.ImagePullPolicy,
			Env:             envVars,
			VolumeMounts:    make([]corev1.VolumeMount, 0),
			ReadinessProbe: &corev1.Probe{
				ProbeHandler: corev1.ProbeHandler{
					HTTPGet: &corev1.HTTPGetAction{
						Port: intstr.FromInt(80),
					},
				},
			},
		}

		if jitsi.Spec.Web.Resources != nil {
			container.Resources = *jitsi.Spec.Web.Resources
		}

		dep.Spec.Template.Spec.Volumes = make([]corev1.Volume, 0)

		if jitsi.Spec.Web.CustomConfig != nil {
			dep.Spec.Template.Spec.Volumes = append(dep.Spec.Template.Spec.Volumes,
				corev1.Volume{
					Name: "custom-config",
					VolumeSource: corev1.VolumeSource{
						ConfigMap: &corev1.ConfigMapVolumeSource{
							LocalObjectReference: *jitsi.Spec.Web.CustomConfig,
							Items: []corev1.KeyToPath{
								{
									Key:  "custom-config.js",
									Path: "custom-config.js",
								},
							},
						},
					},
				})
			container.VolumeMounts = append(container.VolumeMounts, corev1.VolumeMount{
				Name:      "custom-config",
				MountPath: "/config/custom-config.js",
				SubPath:   "custom-config.js",
			})
		}

		if jitsi.Spec.Web.CustomInterfaceConfig != nil {
			dep.Spec.Template.Spec.Volumes = append(dep.Spec.Template.Spec.Volumes,
				corev1.Volume{
					Name: "custom-interface-config",
					VolumeSource: corev1.VolumeSource{
						ConfigMap: &corev1.ConfigMapVolumeSource{
							LocalObjectReference: *jitsi.Spec.Web.CustomInterfaceConfig,
							Items: []corev1.KeyToPath{
								{
									Key:  "custom-interface_config.js",
									Path: "custom-interface_config.js",
								},
							},
						},
					},
				})
			container.VolumeMounts = append(container.VolumeMounts, corev1.VolumeMount{
				Name:      "custom-interface-config",
				MountPath: "/config/custom-interface_config.js",
				SubPath:   "custom-interface_config.js",
			})
		}
		if jitsi.Spec.Web.CustomTitleConfig != nil {
			dep.Spec.Template.Spec.Volumes = append(dep.Spec.Template.Spec.Volumes,
				corev1.Volume{
					Name: "custom-title",
					VolumeSource: corev1.VolumeSource{
						ConfigMap: &corev1.ConfigMapVolumeSource{
							LocalObjectReference: *jitsi.Spec.Web.CustomTitleConfig,
							Items: []corev1.KeyToPath{
								{
									Key:  "custom-title.html",
									Path: "custom-title.html",
								},
							},
						},
					},
				})
			container.VolumeMounts = append(container.VolumeMounts, corev1.VolumeMount{
				Name:      "custom-title",
				MountPath: "/usr/share/jitsi-meet/title.html",
				SubPath:   "custom-title.html",
			})
		}

		if jitsi.Spec.Web.CustomBodyConfig != nil {
			dep.Spec.Template.Spec.Volumes = append(dep.Spec.Template.Spec.Volumes,
				corev1.Volume{
					Name: "custom-body",
					VolumeSource: corev1.VolumeSource{
						ConfigMap: &corev1.ConfigMapVolumeSource{
							LocalObjectReference: *jitsi.Spec.Web.CustomBodyConfig,
							Items: []corev1.KeyToPath{
								{
									Key:  "custom-body.html",
									Path: "custom-body.html",
								},
							},
						},
					},
				})
			container.VolumeMounts = append(container.VolumeMounts, corev1.VolumeMount{
				Name:      "custom-body",
				MountPath: "/usr/share/jitsi-meet/body.html",
				SubPath:   "custom-body.html",
			})
		}
		if jitsi.Spec.Web.CustomCloseConfig != nil {
			dep.Spec.Template.Spec.Volumes = append(dep.Spec.Template.Spec.Volumes,
				corev1.Volume{
					Name: "custom-close",
					VolumeSource: corev1.VolumeSource{
						ConfigMap: &corev1.ConfigMapVolumeSource{
							LocalObjectReference: *jitsi.Spec.Web.CustomCloseConfig,
							Items: []corev1.KeyToPath{
								{
									Key:  "custom-close.html",
									Path: "custom-close.html",
								},
							},
						},
					},
				})
			container.VolumeMounts = append(container.VolumeMounts, corev1.VolumeMount{
				Name:      "custom-close",
				MountPath: "/usr/share/jitsi-meet/static/close3.html",
				SubPath:   "custom-close.html",
			})
		}
		dep.Spec.Template.Spec.Containers = []corev1.Container{container}

		return nil
	})

}
