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

		envVars := jitsi.EnvVars(WebVariables)

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
