package v1alpha1

import (
	"fmt"
	"regexp"
	"strconv"

	appsv1 "k8s.io/api/apps/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2beta2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

var Version = "master"

var defaultEnvVarMap = map[string]string{
	"JVB_AUTH_USER":                        "jvb",
	"JIBRI_RECORDER_USER":                  "recorder",
	"JIBRI_XMPP_USER":                      "jibri",
	"JVB_BREWERY_MUC":                      "jvbbrewery",
	"JIBRI_BREWERY_MUC":                    "jibribrewery",
	"JIBRI_RECORDING_DIR":                  "/config/recordings",
	"JIBRI_PENDING_TIMEOUT":                "90",
	"JIBRI_STRIP_DOMAIN_JID":               "muc",
	"JIBRI_LOGS_DIR":                       "/config/logs",
	"JIBRI_FINALIZE_RECORDING_SCRIPT_PATH": "/config/finalize.sh",
	"ENABLE_RECORDING":                     "1",
	"OCTO_BRIDGE_SELECTION_STRATEGY":       "RegionBasedBridgeSelectionStrategy",
	"TESTING_OCTO_PROBABILITY":             "1",
	"ENABLE_OCTO":                          "1",
	"COLIBRI_REST_ENABLED":                 "1",
	"JVB_STUN_SERVERS":                     "meet-jit-si-turnrelay.jitsi.net:443",
	"DISPLAY":                              ":0",
	"DEPLOYMENTINFO_SHARD":                 "shard",
	"VIDEOQUALITY_BITRATE_H264_LOW":        "200000",
	"VIDEOQUALITY_BITRATE_H264_STANDARD":   "500000",
	"VIDEOQUALITY_BITRATE_H264_HIGH":       "1500000",
	"VIDEOQUALITY_BITRATE_VP8_LOW":         "200000",
	"VIDEOQUALITY_BITRATE_VP8_STANDARD":    "500000",
	"VIDEOQUALITY_BITRATE_VP8_HIGH":        "1500000",
	"VIDEOQUALITY_BITRATE_VP9_LOW":         "100000",
	"VIDEOQUALITY_BITRATE_VP9_STANDARD":    "300000",
	"VIDEOQUALITY_BITRATE_VP9_HIGH":        "1200000",
	"VIDEOQUALITY_PREFERRED_CODEC":         "VP9",
	"JICOFO_ENABLE_HEALTH_CHECKS":          "1",
	"JICOFO_ENABLE_BRIDGE_HEALTH_CHECKS":   "1",
	"JVB_ADVERTISE_PRIVATE_CANDIDATES":     "1",
	"ENABLE_COLIBRI_WEBSOCKET":             "1",
	"NGINX_RESOLVER":                       "kube-dns.kube-system.svc.cluster.local",
	// "DISABLE_HTTPS":                  "1",
	// "ENABLE_HSTS":                    "0",
}

func (jitsi *Jitsi) EnvVarValue(name string) string {
	var value string

	switch name {
	case "TZ":
		value = jitsi.Spec.Timezone
	case "XMPP_SERVER":
		value = fmt.Sprintf("%s-prosody", jitsi.Name)
	case "XMPP_BOSH_URL_BASE":
		value = fmt.Sprintf("http://%s.%s.svc.cluster.local:5280", jitsi.EnvVarValue("XMPP_SERVER"), jitsi.Namespace)
	case "JVB_PORT":
		value = strconv.FormatInt(int64(*jitsi.Spec.JVB.Ports.UDP), 10)
	case "JVB_TCP_PORT":
		value = strconv.FormatInt(int64(*jitsi.Spec.JVB.Ports.TCP), 10)
	case "DEPLOYMENTINFO_USERREGION":
		value = jitsi.Spec.Region
	case "JVB_OCTO_REGION":
		value = jitsi.Spec.Region
	case "DEPLOYMENTINFO_REGION":
		value = jitsi.Spec.Region
	case "PUBLIC_URL":
		value = "https://" + jitsi.Spec.Domain
	case "SHUTDOWN_REST_ENABLED":
		if jitsi.Spec.JVB.GracefulShutdown || jitsi.Spec.Variables["SHUTDOWN_REST_ENABLED"] == "1" {
			value = "1"
		} else {
			value = "0"
		}
	default:
		if jitsi.Spec.Variables[name] != "" {
			value = jitsi.Spec.Variables[name]
		} else {
			value = defaultEnvVarMap[name]
		}
	}

	return value
}

func (jitsi *Jitsi) EnvVar(name string) corev1.EnvVar {
	return corev1.EnvVar{
		Name:  name,
		Value: jitsi.EnvVarValue(name),
	}
}

func (jitsi *Jitsi) EnvVars(names []string) []corev1.EnvVar {
	var envVars []corev1.EnvVar

	for _, env := range names {
		if len(jitsi.EnvVar(env).Value) > 0 {
			envVars = append(envVars, jitsi.EnvVar(env))
		}
	}
	return envVars
}

func (jitsi *Jitsi) SetDefaults() {
	if len(jitsi.Spec.Image.Registry) == 0 {
		jitsi.Spec.Image.Registry = "ghcr.io/jitsi-contrib/jitsi-kubernetes-operator"
	}

	if len(jitsi.Spec.Image.Tag) == 0 {
		jitsi.Spec.Image.Tag = Version
	}

	if len(jitsi.Spec.Image.PullPolicy) == 0 {
		if ok, _ := regexp.MatchString(`\d+\.\d+\.\d+`, jitsi.Spec.Image.Tag); ok {
			jitsi.Spec.Image.PullPolicy = corev1.PullIfNotPresent
		} else {
			jitsi.Spec.Image.PullPolicy = corev1.PullAlways
		}
	}

	if jitsi.Spec.JVB.Strategy.Replicas == nil {
		defaultReplicas := int32(1)
		jitsi.Spec.JVB.Strategy.Replicas = &defaultReplicas
	}

	if len(jitsi.Spec.JVB.Strategy.Type) == 0 {
		jitsi.Spec.JVB.Strategy.Type = JVBStrategyStatic
	}

	if jitsi.Spec.JVB.Ports.TCP == nil {
		defaultPort := int32(30301)
		jitsi.Spec.JVB.Ports.TCP = &defaultPort
	}

	if jitsi.Spec.JVB.Ports.UDP == nil {
		defaultPort := int32(10000)
		jitsi.Spec.JVB.Ports.UDP = &defaultPort
	}

	if jitsi.Spec.JVB.ContainerRuntime == nil {
		jitsi.Spec.JVB.ContainerRuntime = &ContainerRuntime{}
	}

	if len(jitsi.Spec.JVB.Image) == 0 {
		jitsi.Spec.JVB.Image = fmt.Sprintf("%s/jvb:%s", jitsi.Spec.Image.Registry, jitsi.Spec.Image.Tag)
	}

	if len(jitsi.Spec.JVB.ImagePullPolicy) == 0 {
		jitsi.Spec.JVB.ImagePullPolicy = jitsi.Spec.Image.PullPolicy
	}

	if jitsi.Spec.Jibri.ContainerRuntime == nil {
		jitsi.Spec.Jibri.ContainerRuntime = &ContainerRuntime{}
	}

	if jitsi.Spec.Jibri.Enabled {
		if jitsi.Spec.Jibri.Replicas == nil {
			defaultReplicas := int32(1)
			jitsi.Spec.Jibri.Replicas = &defaultReplicas
		}

		if jitsi.Spec.Jibri.ContainerRuntime == nil {
			jitsi.Spec.Jibri.ContainerRuntime = &ContainerRuntime{}
		}

		if len(jitsi.Spec.Jibri.Image) == 0 {
			jitsi.Spec.Jibri.Image = fmt.Sprintf("%s/jibri:%s", jitsi.Spec.Image.Registry, jitsi.Spec.Image.Tag)
		}

		if len(jitsi.Spec.Jibri.ImagePullPolicy) == 0 {
			jitsi.Spec.Jibri.ImagePullPolicy = jitsi.Spec.Image.PullPolicy
		}
	}

	if jitsi.Spec.Prosody.ContainerRuntime == nil {
		jitsi.Spec.Prosody.ContainerRuntime = &ContainerRuntime{}
	}

	if len(jitsi.Spec.Prosody.Image) == 0 {
		jitsi.Spec.Prosody.Image = fmt.Sprintf("%s/prosody:%s", jitsi.Spec.Image.Registry, jitsi.Spec.Image.Tag)
	}

	if len(jitsi.Spec.Prosody.ImagePullPolicy) == 0 {
		jitsi.Spec.Prosody.ImagePullPolicy = jitsi.Spec.Image.PullPolicy
	}

	if jitsi.Spec.Jicofo.ContainerRuntime == nil {
		jitsi.Spec.Jicofo.ContainerRuntime = &ContainerRuntime{}
	}

	if len(jitsi.Spec.Jicofo.Image) == 0 {
		jitsi.Spec.Jicofo.Image = fmt.Sprintf("%s/jicofo:%s", jitsi.Spec.Image.Registry, jitsi.Spec.Image.Tag)
	}

	if len(jitsi.Spec.Jicofo.ImagePullPolicy) == 0 {
		jitsi.Spec.Jicofo.ImagePullPolicy = jitsi.Spec.Image.PullPolicy
	}

	if jitsi.Spec.Web.Replicas == nil {
		defaultReplicas := int32(1)
		jitsi.Spec.Web.Replicas = &defaultReplicas
	}

	if jitsi.Spec.Web.ContainerRuntime == nil {
		jitsi.Spec.Web.ContainerRuntime = &ContainerRuntime{}
	}

	if len(jitsi.Spec.Web.Image) == 0 {
		jitsi.Spec.Web.Image = fmt.Sprintf("%s/web:%s", jitsi.Spec.Image.Registry, jitsi.Spec.Image.Tag)
	}

	if len(jitsi.Spec.Web.ImagePullPolicy) == 0 {
		jitsi.Spec.Web.ImagePullPolicy = jitsi.Spec.Image.PullPolicy
	}

	if jitsi.Spec.Ingress.Annotations == nil {
		jitsi.Spec.Ingress.Annotations = make(map[string]string)
	}
}

func (jitsi *Jitsi) ComponentLabels(component string) labels.Set {
	l := jitsi.Labels()
	l["app.kubernetes.io/component"] = component

	return l
}

func (jitsi *Jitsi) Labels() labels.Set {
	labels := labels.Set{
		"app.kubernetes.io/name":       "jitsi",
		"app.kubernetes.io/part-of":    "jitsi",
		"app.kubernetes.io/instance":   jitsi.ObjectMeta.Name,
		"app.kubernetes.io/managed-by": "jitsi-operator",
	}

	return labels
}

func (jitsi *Jitsi) JibriDeployment() appsv1.Deployment {
	return appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-jibri", jitsi.Name),
			Namespace: jitsi.Namespace,
		},
	}
}

func (jitsi *Jitsi) JVBHPA() autoscalingv2.HorizontalPodAutoscaler {
	return autoscalingv2.HorizontalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-jvb", jitsi.Name),
			Namespace: jitsi.Namespace,
		},
	}
}

func (jitsi *Jitsi) JVBDeployment() appsv1.Deployment {
	return appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-jvb", jitsi.Name),
			Namespace: jitsi.Namespace,
		},
	}
}

func (jitsi *Jitsi) JVBDaemonSet() appsv1.DaemonSet {
	return appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-jvb", jitsi.Name),
			Namespace: jitsi.Namespace,
		},
	}
}
