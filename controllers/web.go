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
		// 	dep.Spec.Replicas = 1
		dep.Spec.Strategy.Type = appsv1.RecreateDeploymentStrategyType
		dep.Spec.Template.Spec.Containers = []corev1.Container{
			{
				Name:  "web",
				Image: "jitsi/web",
				Env: []corev1.EnvVar{
					jitsi.EnvVar("ENABLE_COLIBRI_WEBSOCKET"),
					jitsi.EnvVar("ENABLE_FLOC"),
					jitsi.EnvVar("ENABLE_LETSENCRYPT"),
					jitsi.EnvVar("ENABLE_HTTP_REDIRECT"),
					jitsi.EnvVar("ENABLE_HSTS"),
					jitsi.EnvVar("ENABLE_XMPP_WEBSOCKET"),
					jitsi.EnvVar("DISABLE_HTTPS"),
					jitsi.EnvVar("DISABLE_DEEP_LINKING"),
					jitsi.EnvVar("LETSENCRYPT_DOMAIN"),
					jitsi.EnvVar("LETSENCRYPT_EMAIL"),
					jitsi.EnvVar("LETSENCRYPT_USE_STAGING"),
					jitsi.EnvVar("PUBLIC_URL"),
					jitsi.EnvVar("TZ"),
					jitsi.EnvVar("AMPLITUDE_ID"),
					jitsi.EnvVar("ANALYTICS_SCRIPT_URLS"),
					jitsi.EnvVar("ANALYTICS_WHITELISTED_EVENTS"),
					jitsi.EnvVar("CALLSTATS_CUSTOM_SCRIPT_URL"),
					jitsi.EnvVar("CALLSTATS_ID"),
					jitsi.EnvVar("CALLSTATS_SECRET"),
					jitsi.EnvVar("CHROME_EXTENSION_BANNER_JSON"),
					jitsi.EnvVar("CONFCODE_URL"),
					jitsi.EnvVar("CONFIG_EXTERNAL_CONNECT"),
					jitsi.EnvVar("DEFAULT_LANGUAGE"),
					jitsi.EnvVar("DEPLOYMENTINFO_ENVIRONMENT"),
					jitsi.EnvVar("DEPLOYMENTINFO_ENVIRONMENT_TYPE"),
					jitsi.EnvVar("DEPLOYMENTINFO_REGION"),
					jitsi.EnvVar("DEPLOYMENTINFO_SHARD"),
					jitsi.EnvVar("DEPLOYMENTINFO_USERREGION"),
					jitsi.EnvVar("DIALIN_NUMBERS_URL"),
					jitsi.EnvVar("DIALOUT_AUTH_URL"),
					jitsi.EnvVar("DIALOUT_CODES_URL"),
					jitsi.EnvVar("DROPBOX_APPKEY"),
					jitsi.EnvVar("DROPBOX_REDIRECT_URI"),
					jitsi.EnvVar("DYNAMIC_BRANDING_URL"),
					jitsi.EnvVar("ENABLE_AUDIO_PROCESSING"),
					jitsi.EnvVar("ENABLE_AUTH"),
					jitsi.EnvVar("ENABLE_CALENDAR"),
					jitsi.EnvVar("ENABLE_FILE_RECORDING_SERVICE"),
					jitsi.EnvVar("ENABLE_FILE_RECORDING_SERVICE_SHARING"),
					jitsi.EnvVar("ENABLE_GUESTS"),
					jitsi.EnvVar("ENABLE_IPV6"),
					jitsi.EnvVar("ENABLE_LIPSYNC"),
					jitsi.EnvVar("ENABLE_NO_AUDIO_DETECTION"),
					jitsi.EnvVar("ENABLE_P2P"),
					jitsi.EnvVar("ENABLE_PREJOIN_PAGE"),
					jitsi.EnvVar("ENABLE_WELCOME_PAGE"),
					jitsi.EnvVar("ENABLE_CLOSE_PAGE"),
					jitsi.EnvVar("ENABLE_RECORDING"),
					jitsi.EnvVar("ENABLE_REMB"),
					jitsi.EnvVar("ENABLE_REQUIRE_DISPLAY_NAME"),
					jitsi.EnvVar("ENABLE_SIMULCAST"),
					jitsi.EnvVar("ENABLE_STATS_ID"),
					jitsi.EnvVar("ENABLE_STEREO"),
					jitsi.EnvVar("ENABLE_SUBDOMAINS"),
					jitsi.EnvVar("ENABLE_TALK_WHILE_MUTED"),
					jitsi.EnvVar("ENABLE_TCC"),
					jitsi.EnvVar("ENABLE_TRANSCRIPTIONS"),
					jitsi.EnvVar("ETHERPAD_PUBLIC_URL"),
					jitsi.EnvVar("ETHERPAD_URL_BASE"),
					jitsi.EnvVar("GOOGLE_ANALYTICS_ID"),
					jitsi.EnvVar("GOOGLE_API_APP_CLIENT_ID"),
					jitsi.EnvVar("INVITE_SERVICE_URL"),
					jitsi.EnvVar("JICOFO_AUTH_USER"),
					jitsi.EnvVar("MATOMO_ENDPOINT"),
					jitsi.EnvVar("MATOMO_SITE_ID"),
					jitsi.EnvVar("MICROSOFT_API_APP_CLIENT_ID"),
					jitsi.EnvVar("NGINX_RESOLVER"),
					jitsi.EnvVar("NGINX_WORKER_PROCESSES"),
					jitsi.EnvVar("NGINX_WORKER_CONNECTIONS"),
					jitsi.EnvVar("PEOPLE_SEARCH_URL"),
					jitsi.EnvVar("RESOLUTION"),
					jitsi.EnvVar("RESOLUTION_MIN"),
					jitsi.EnvVar("RESOLUTION_WIDTH"),
					jitsi.EnvVar("RESOLUTION_WIDTH_MIN"),
					jitsi.EnvVar("START_AUDIO_ONLY"),
					jitsi.EnvVar("START_AUDIO_MUTED"),
					jitsi.EnvVar("START_WITH_AUDIO_MUTED"),
					jitsi.EnvVar("START_SILENT"),
					jitsi.EnvVar("DISABLE_AUDIO_LEVELS"),
					jitsi.EnvVar("ENABLE_NOISY_MIC_DETECTION"),
					jitsi.EnvVar("START_BITRATE"),
					jitsi.EnvVar("DESKTOP_SHARING_FRAMERATE_MIN"),
					jitsi.EnvVar("DESKTOP_SHARING_FRAMERATE_MAX"),
					jitsi.EnvVar("START_VIDEO_MUTED"),
					jitsi.EnvVar("START_WITH_VIDEO_MUTED"),
					jitsi.EnvVar("TESTING_CAP_SCREENSHARE_BITRATE"),
					jitsi.EnvVar("TESTING_OCTO_PROBABILITY"),
					jitsi.EnvVar("XMPP_AUTH_DOMAIN"),
					jitsi.EnvVar("XMPP_BOSH_URL_BASE"),
					jitsi.EnvVar("XMPP_DOMAIN"),
					jitsi.EnvVar("XMPP_GUEST_DOMAIN"),
					jitsi.EnvVar("XMPP_MUC_DOMAIN"),
					jitsi.EnvVar("XMPP_RECORDER_DOMAIN"),
					jitsi.EnvVar("TOKEN_AUTH_URL"),
				},
			},
		}
		return nil
	})

}
