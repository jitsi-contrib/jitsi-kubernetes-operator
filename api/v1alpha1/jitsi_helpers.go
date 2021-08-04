package v1alpha1

import (
	"fmt"
	"strconv"

	appsv1 "k8s.io/api/apps/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2beta2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/intstr"
)

var defaultEnvVarMap = map[string]string{
	"JICOFO_AUTH_USER":               "focus",
	"JVB_AUTH_USER":                  "jvb",
	"JIBRI_RECORDER_USER":            "recorder",
	"JIBRI_XMPP_USER":                "jibri",
	"JVB_BREWERY_MUC":                "jvbbrewery",
	"JIBRI_BREWERY_MUC":              "jibribrewery",
	"JIBRI_RECORDING_DIR":            "/config/recordings",
	"JIBRI_PENDING_TIMEOUT":          "90",
	"JIBRI_STRIP_DOMAIN_JID":         "muc",
	"JIBRI_LOGS_DIR":                 "/config/logs",
	"ENABLE_RECORDING":               "1",
	"OCTO_BRIDGE_SELECTION_STRATEGY": "RegionBasedBridgeSelectionStrategy",
	"TESTING_OCTO_PROBABILITY":       "1",
	"ENABLE_OCTO":                    "1",
	"JVB_ENABLE_APIS":                "rest,colibri",
	"JVB_STUN_SERVERS":               "stun2.l.google.com:19302",
	"DISPLAY":                        ":0",
	// "DISABLE_HTTPS":                  "1",
	// "ENABLE_HSTS":                    "0",
}

var jvbEnvs = []string{
	"ENABLE_COLIBRI_WEBSOCKET",
	"ENABLE_OCTO",
	"DOCKER_HOST_ADDRESS",
	"XMPP_AUTH_DOMAIN",
	"XMPP_INTERNAL_MUC_DOMAIN",
	"XMPP_SERVER",
	"JVB_AUTH_USER",
	"JVB_AUTH_PASSWORD",
	"JVB_BREWERY_MUC",
	"JVB_PORT",
	"JVB_TCP_HARVESTER_DISABLED",
	"JVB_TCP_PORT",
	"JVB_TCP_MAPPED_PORT",
	"JVB_STUN_SERVERS",
	"JVB_ENABLE_APIS",
	"JVB_WS_DOMAIN",
	"JVB_WS_SERVER_ID",
	"PUBLIC_URL",
	"JVB_OCTO_BIND_ADDRESS",
	"JVB_OCTO_PUBLIC_ADDRESS",
	"JVB_OCTO_BIND_PORT",
	"JVB_OCTO_REGION",
	"TZ",
}

func (jitsi *Jitsi) EnvVarValue(name string) string {
	var value string

	switch name {
	case "TZ":
		value = jitsi.Spec.Timezone
	case "XMPP_SERVER":
		value = fmt.Sprintf("%s-prosody", jitsi.Name)
	case "XMPP_BOSH_URL_BASE":
		value = "http://" + jitsi.EnvVarValue("XMPP_SERVER") + ":5280"
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
	case "XMPP_DOMAIN":
		value = jitsi.Spec.Domain
	case "XMPP_AUTH_DOMAIN":
		value = "auth." + jitsi.EnvVarValue("XMPP_DOMAIN")
	case "XMPP_INTERNAL_MUC_DOMAIN":
		value = "internal-muc." + jitsi.EnvVarValue("XMPP_DOMAIN")
	case "XMPP_MUC_DOMAIN":
		value = "muc." + jitsi.EnvVarValue("XMPP_DOMAIN")
	case "XMPP_RECORDER_DOMAIN":
		value = "recorder." + jitsi.EnvVarValue("XMPP_DOMAIN")
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

func (jitsi *Jitsi) SetDefaults() {
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
		defaultPort := int32(30300)
		jitsi.Spec.JVB.Ports.UDP = &defaultPort
	}

	if len(jitsi.Spec.Version.Channel) == 0 {
		jitsi.Spec.Version.Channel = VersionStable
	}

	if len(jitsi.Spec.Version.Tag) == 0 {
		jitsi.Spec.Version.Tag = "latest"
	}

	if jitsi.Spec.JVB.ContainerRuntime == nil {
		jitsi.Spec.JVB.ContainerRuntime = &ContainerRuntime{}
	}

	if len(jitsi.Spec.JVB.Image) == 0 {
		if jitsi.Spec.Version.Tag == "latest" {
			jitsi.Spec.JVB.Image = "jitsi/jvb:latest"
		} else {
			jitsi.Spec.JVB.ContainerRuntime.Image = fmt.Sprintf("jitsi/jvb:%s-%s", jitsi.Spec.Version.Channel, jitsi.Spec.Version.Tag)
		}
	}

	if len(jitsi.Spec.JVB.ImagePullPolicy) == 0 {
		jitsi.Spec.JVB.ImagePullPolicy = corev1.PullIfNotPresent
	}

	if jitsi.Spec.Jibri.ContainerRuntime == nil {
		jitsi.Spec.JVB.ContainerRuntime = &ContainerRuntime{}
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
			if jitsi.Spec.Version.Tag == "latest" {
				jitsi.Spec.Jibri.Image = "jitsi/jibri:latest"
			} else {
				jitsi.Spec.Jibri.ContainerRuntime.Image = fmt.Sprintf("jitsi/jibri:%s-%s", jitsi.Spec.Version.Channel, jitsi.Spec.Version.Tag)
			}
		}

		if len(jitsi.Spec.Jibri.ImagePullPolicy) == 0 {
			jitsi.Spec.Jibri.ImagePullPolicy = corev1.PullIfNotPresent
		}
	}

	if jitsi.Spec.Prosody.ContainerRuntime == nil {
		jitsi.Spec.Prosody.ContainerRuntime = &ContainerRuntime{}
	}

	if len(jitsi.Spec.Prosody.Image) == 0 {
		if jitsi.Spec.Version.Tag == "latest" {
			jitsi.Spec.Prosody.Image = "jitsi/prosody:latest"
		} else {
			jitsi.Spec.Prosody.ContainerRuntime.Image = fmt.Sprintf("jitsi/prosody:%s-%s", jitsi.Spec.Version.Channel, jitsi.Spec.Version.Tag)
		}
	}

	if len(jitsi.Spec.Prosody.ImagePullPolicy) == 0 {
		jitsi.Spec.Prosody.ImagePullPolicy = corev1.PullIfNotPresent
	}

	if jitsi.Spec.Jicofo.ContainerRuntime == nil {
		jitsi.Spec.Jicofo.ContainerRuntime = &ContainerRuntime{}
	}

	if len(jitsi.Spec.Jicofo.Image) == 0 {
		if jitsi.Spec.Version.Tag == "latest" {
			jitsi.Spec.Jicofo.Image = "jitsi/jicofo:latest"
		} else {
			jitsi.Spec.Jicofo.ContainerRuntime.Image = fmt.Sprintf("jitsi/jicofo:%s-%s", jitsi.Spec.Version.Channel, jitsi.Spec.Version.Tag)
		}
	}

	if len(jitsi.Spec.Jicofo.ImagePullPolicy) == 0 {
		jitsi.Spec.Jicofo.ImagePullPolicy = corev1.PullIfNotPresent
	}

	if jitsi.Spec.Web.Replicas == nil {
		defaultReplicas := int32(1)
		jitsi.Spec.Web.Replicas = &defaultReplicas
	}

	if jitsi.Spec.Web.ContainerRuntime == nil {
		jitsi.Spec.Web.ContainerRuntime = &ContainerRuntime{}
	}

	if len(jitsi.Spec.Web.Image) == 0 {
		if jitsi.Spec.Version.Tag == "latest" {
			jitsi.Spec.Web.Image = "jitsi/web:latest"
		} else {
			jitsi.Spec.Web.ContainerRuntime.Image = fmt.Sprintf("jitsi/web:%s-%s", jitsi.Spec.Version.Channel, jitsi.Spec.Version.Tag)
		}
	}

	if len(jitsi.Spec.Web.ImagePullPolicy) == 0 {
		jitsi.Spec.Web.ImagePullPolicy = corev1.PullIfNotPresent
	}
}

func (jitsi *Jitsi) JVBPodTemplateSpec(podSpec *corev1.PodTemplateSpec) {
	podSpec.Spec.Volumes = []corev1.Volume{
		{
			Name: "jvb-config",
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: fmt.Sprintf("%s-jvb", jitsi.Name),
					},
					Items: []corev1.KeyToPath{
						{
							Key:  "sip-communicator.properties",
							Path: "sip-communicator.properties",
						},
					},
				},
			},
		},
	}

	envVars := []corev1.EnvVar{
		{
			Name: "LOCAL_ADDRESS",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					FieldPath: "status.podIP",
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
		// TODO options to manage host IP through vars or stun servers
		// {
		// 	Name: "DOCKER_HOST_ADDRESS",
		// 	ValueFrom: &corev1.EnvVarSource{
		// 		FieldRef: &corev1.ObjectFieldSelector{
		// 			FieldPath: "status.hostIP",
		// 		},
		// 	},
		// },

		// {
		// Default 0.0.0.0
		// 	Name: "JVB_OCTO_BIND_ADDRESS",
		// 	ValueFrom: &corev1.EnvVarSource{
		// 		FieldRef: &corev1.ObjectFieldSelector{
		// 			FieldPath: "status.podIP",
		// 		},
		// 	},
		// },

		{
			Name: "JVB_OCTO_PUBLIC_ADDRESS",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					FieldPath: "status.podIP",
				},
			},
		},
	}

	for _, env := range jvbEnvs {
		if len(jitsi.EnvVar(env).Value) > 0 {
			envVars = append(envVars, jitsi.EnvVar(env))
		}
	}

	jvbContainer := corev1.Container{
		Name:            "jvb",
		Image:           "jitsi/jvb",
		ImagePullPolicy: "Always",
		Env:             envVars,
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      "jvb-config",
				MountPath: "/defaults/sip-communicator.properties",
				SubPath:   "sip-communicator.properties",
				ReadOnly:  true,
			},
		},
		ReadinessProbe: &corev1.Probe{
			Handler: corev1.Handler{
				HTTPGet: &corev1.HTTPGetAction{
					Path: "/about/health",
					Port: intstr.FromInt(8080),
				},
			},
			InitialDelaySeconds: 10,
		},
	}

	if jitsi.Spec.JVB.GracefulShutdown {
		jvbContainer.Lifecycle = &corev1.Lifecycle{
			PreStop: &corev1.Handler{
				Exec: &corev1.ExecAction{
					Command: []string{
						"bash", "-c", "/usr/share/jitsi-videobridge/graceful_shutdown.sh -p $(s6-svstat -o pid /var/run/s6/services/jvb) -t 3 -s",
					},
				},
			},
		}
	}

	if jitsi.Spec.JVB.Resources != nil {
		jvbContainer.Resources = *jitsi.Spec.JVB.Resources
	}

	podSpec.Spec.Containers = []corev1.Container{jvbContainer}
}

func (jitsi *Jitsi) ComponentLabels(component string) labels.Set {
	l := jitsi.Labels()
	l["app.kubernetes.io/component"] = component

	return l
}

func (jitsi *Jitsi) Labels() labels.Set {
	labels := labels.Set{
		"app.kubernetes.io/name":       "jitsi",
		"app.kubernetes.io/part-of":    "jistsi",
		"app.kubernetes.io/instance":   jitsi.ObjectMeta.Name,
		"app.kubernetes.io/managed-by": "jitsi-operator",
		"app.kubernetes.io/version":    fmt.Sprintf("%s-%s", jitsi.Spec.Version.Channel, jitsi.Spec.Version.Tag),
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
