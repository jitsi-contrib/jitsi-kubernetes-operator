package controllers

import (
	"fmt"
	"jitsi-operator/api/v1alpha1"

	"github.com/presslabs/controller-util/syncer"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func NewJVBPodMonitorSyncer(jitsi *v1alpha1.Jitsi, c client.Client) syncer.Interface {
	mon := &monitoringv1.PodMonitor{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-jvb", jitsi.Name),
			Namespace: jitsi.Namespace,
		},
	}

	return syncer.NewObjectSyncer("PodMonitor", jitsi, mon, c, func() error {
		mon.Labels = jitsi.ComponentLabels("jvb")

		mon.Spec.Selector = metav1.LabelSelector{
			MatchLabels: jitsi.ComponentLabels("jvb"),
		}
		mon.Spec.PodMetricsEndpoints = []monitoringv1.PodMetricsEndpoint{
			{
				Port: "metrics",
			},
		}

		return nil
	})

}

func NewJicofoServiceMonitorSyncer(jitsi *v1alpha1.Jitsi, c client.Client) syncer.Interface {
	mon := &monitoringv1.PodMonitor{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-jicofo", jitsi.Name),
			Namespace: jitsi.Namespace,
		},
	}

	return syncer.NewObjectSyncer("PodMonitor", jitsi, mon, c, func() error {
		mon.Labels = jitsi.ComponentLabels("jicofo")

		mon.Spec.Selector = metav1.LabelSelector{
			MatchLabels: jitsi.ComponentLabels("jicofo"),
		}
		mon.Spec.PodMetricsEndpoints = []monitoringv1.PodMetricsEndpoint{
			{
				Port: "metrics",
			},
		}

		return nil
	})

}

func NewMetricsContainer(component string) corev1.Container {
	container := corev1.Container{
		Name: "metrics",
		//	Image: "libresh/jitsi-exporter:latest",
		Image: "unteem/jitsi-exporter:t02",
		Ports: []corev1.ContainerPort{
			{
				Name:          "metrics",
				ContainerPort: 9210,
			},
		},
	}

	if component == "jvb" {
		container.Args = []string{"-target.name=jvb", "-target.url=http://127.0.0.1:8080/colibri/stats"}
	}

	if component == "jicofo" {
		container.Args = []string{"-target.name=jicofo", "-target.url=http://127.0.0.1:8888/stats"}
	}

	return container
}
