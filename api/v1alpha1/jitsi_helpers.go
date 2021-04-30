package v1alpha1

import (
	"fmt"
	"strconv"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
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
	"JVB_STUN_SERVERS":               "meet-jit-si-turnrelay.jitsi.net:443",
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
		value = "http://" + jitsi.EnvVarValue("XMPP_SERVER") + ":5280"
	case "JVB_PORT":
		value = strconv.FormatInt(int64(*jitsi.Spec.JVB.Ports.TCP), 10)
	case "JVB_TCP_PORT":
		value = strconv.FormatInt(int64(*jitsi.Spec.JVB.Ports.UDP), 10)
	case "DEPLOYMENTINFO_USERREGION":
		value = jitsi.Spec.Region
	case "JVB_OCTO_REGION":
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
		value = defaultEnvVarMap[name]
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

	if jitsi.Spec.JVB.Ports.TCP == nil {
		defaultPort := int32(30300)
		jitsi.Spec.JVB.Ports.TCP = &defaultPort
	}

	if jitsi.Spec.JVB.Ports.UDP == nil {
		defaultPort := int32(30301)
		jitsi.Spec.JVB.Ports.UDP = &defaultPort
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

	podSpec.Spec.Containers = []corev1.Container{
		{
			Name:            "jvb",
			Image:           "jitsi/jvb",
			ImagePullPolicy: "Always",
			Env: []corev1.EnvVar{
				{
					Name: "LOCAL_ADDRESS",
					ValueFrom: &corev1.EnvVarSource{
						FieldRef: &corev1.ObjectFieldSelector{
							FieldPath: "status.podIP",
						},
					},
				},
				// {
				// 	Name: "DOCKER_HOST_ADDRESS",
				// 	ValueFrom: &corev1.EnvVarSource{
				// 		FieldRef: &corev1.ObjectFieldSelector{
				// 			FieldPath: "status.hostIP",
				// 		},
				// 	},
				// },
				// {
				// 	Name: "JVB_OCTO_BIND_ADDRESS",
				// 	ValueFrom: &corev1.EnvVarSource{
				// 		FieldRef: &corev1.ObjectFieldSelector{
				// 			FieldPath: "status.podIP",
				// 		},
				// 	},
				// },
				// {
				// 	Name: "JVB_OCTO_PUBLIC_ADDRESS",
				// 	ValueFrom: &corev1.EnvVarSource{
				// 		FieldRef: &corev1.ObjectFieldSelector{
				// 			FieldPath: "status.hostIP",
				// 		},
				// 	},
				// },
				// jitsi.EnvVar("ENABLE_COLIBRI_WEBSOCKET"),
				jitsi.EnvVar("ENABLE_OCTO"),
				// jitsi.EnvVar("DOCKER_HOST_ADDRESS"),
				jitsi.EnvVar("XMPP_AUTH_DOMAIN"),
				jitsi.EnvVar("XMPP_INTERNAL_MUC_DOMAIN"),
				jitsi.EnvVar("XMPP_SERVER"),
				jitsi.EnvVar("JVB_AUTH_USER"),
				jitsi.EnvVar("JVB_BREWERY_MUC"),
				jitsi.EnvVar("JVB_PORT"),
				// jitsi.EnvVar("JVB_TCP_HARVESTER_DISABLED"),
				jitsi.EnvVar("JVB_TCP_PORT"),
				// jitsi.EnvVar("JVB_TCP_MAPPED_PORT"),
				jitsi.EnvVar("JVB_STUN_SERVERS"),
				jitsi.EnvVar("JVB_ENABLE_APIS"),
				// jitsi.EnvVar("JVB_WS_DOMAIN"),
				// jitsi.EnvVar("JVB_WS_SERVER_ID"),
				jitsi.EnvVar("PUBLIC_URL"),
				// jitsi.EnvVar("JVB_OCTO_BIND_PORT"),
				jitsi.EnvVar("JVB_OCTO_REGION"),
				jitsi.EnvVar("TZ"),
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
			},
			VolumeMounts: []corev1.VolumeMount{
				{
					Name:      "jvb-config",
					MountPath: "/defaults/sip-communicator.properties",
					SubPath:   "sip-communicator.properties",
					ReadOnly:  true,
				},
			},
		},
	}
}

// func MutateWebService(jitsi v1alpha1.Jitsi, svc corev1.Service) error {
// 	//	if jitsi.JVB.Ports.TCP
// 	svc.Name = fmt.Sprintf("%s-web", jitsi.Name)
// 	svc.Namespace = jitsi.Namespace
// 	svc.Labels = ComponentLabels(web)

// 	svc.Spec.Type = corev1.clusterIP

// 	port := []corev1.ServicePort{
// 		{
// 			Name:       "http",
// 			Port:       80,
// 			TargetPort: 80,
// 			Protocol:   corev1.ProtocolTCP,
// 		},
// 	}

// 	return nil
// }

// func MutatWebDeployment(jitsi v1alpha1.Jitsi, dep corev1.Deployment) error {
// 	dep.Name = fmt.Sprintf("%s-web", jitsi.Name)
// 	dep.Namespace = jitsi.Namespace
// 	dep.Labels = ComponentLabels(web)

// 	dep.Spec.Template.Spec.Containers = []corev1.Container{
// 		{
// 			Name:  "web",
// 			Image: " jitsi/web",
// 		},
// 	}
// 	return nil
// }

func (jitsi *Jitsi) ComponentLabels(component string) labels.Set {
	l := jitsi.Labels()
	l["app.kubernetes.io/component"] = component

	return l
}

func (jitsi *Jitsi) Labels() labels.Set {
	// partOf := "jitsi"

	// if jitsi.ObjectMeta.Labels != nil && len(jitsi.ObjectMeta.Labels["app.kubernetes.io/part-of"]) > 0 {
	// 	partOf = jitsi.ObjectMeta.Labels["app.kubernetes.io/part-of"]
	// }

	labels := labels.Set{
		"app.kubernetes.io/name":       "jitsi",
		"app.kubernetes.io/part-of":    "jistsi",
		"app.kubernetes.io/instance":   jitsi.ObjectMeta.Name,
		"app.kubernetes.io/managed-by": "jitsi-operator",
		"app.kubernetes.io/version":    jitsi.Spec.Version,
	}

	return labels
}
