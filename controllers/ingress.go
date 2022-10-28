package controllers

import (
	"fmt"

	"github.com/jitsi-contrib/jitsi-kubernetes-operator/api/v1alpha1"

	"github.com/presslabs/controller-util/pkg/syncer"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

func NewIngressSyncer(jitsi *v1alpha1.Jitsi, c client.Client) syncer.Interface {
	obj := &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-web", jitsi.Name),
			Namespace: jitsi.Namespace,
		},
	}

	return syncer.NewObjectSyncer("Ingress", jitsi, obj, c, func() error {
		pathType := networkingv1.PathTypePrefix

		obj.Annotations = jitsi.Spec.Ingress.Annotations
		obj.Annotations["nginx.ingress.kubernetes.io/proxy-read-timeout"] = "3600"
		obj.Annotations["nginx.ingress.kubernetes.io/proxy-send-timeout"] = "3600"

		obj.Labels = jitsi.ComponentLabels("web")
		obj.Spec.Rules = []networkingv1.IngressRule{
			{
				Host: jitsi.Spec.Domain,
				IngressRuleValue: networkingv1.IngressRuleValue{
					HTTP: &networkingv1.HTTPIngressRuleValue{
						Paths: []networkingv1.HTTPIngressPath{
							{
								Path:     "/",
								PathType: &pathType,
								Backend: networkingv1.IngressBackend{
									Service: &networkingv1.IngressServiceBackend{
										Name: fmt.Sprintf("%s-web", jitsi.Name),
										Port: networkingv1.ServiceBackendPort{
											Name: "http",
										},
									},
								},
							},
						},
					},
				},
			},
		}

		if jitsi.Spec.Ingress.TLS {
			obj.Spec.TLS = []networkingv1.IngressTLS{
				{
					Hosts:      []string{jitsi.Spec.Domain},
					SecretName: jitsi.Spec.Domain + "-tls",
				},
			}

		}

		return nil
	})

}
