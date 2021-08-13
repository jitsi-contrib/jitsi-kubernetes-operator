package controllers

import (
	"fmt"
	"jitsi-operator/api/v1alpha1"

	"github.com/presslabs/controller-util/syncer"
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

		// with this nginx ingress becomes a requierement
		// TODO make nginx ingress an option
		obj.Annotations["nginx.ingress.kubernetes.io/server-snippet"] = fmt.Sprintf(`add_header X-Jitsi-Shard shard;
			location = /xmpp-websocket {
			    proxy_pass http://%s-prosody.%s:5280/xmpp-websocket;
			    proxy_http_version 1.1;
		
			    proxy_set_header Connection "upgrade";
			    proxy_set_header Upgrade $http_upgrade;
		
			    proxy_set_header Host %s;
			    proxy_set_header X-Forwarded-For $remote_addr;
			    tcp_nodelay on;
			}
			location ~ ^/colibri-ws/([a-zA-Z0-9-\.]+)/(.*) {
				proxy_pass http://$1:9090/colibri-ws/$1/$2$is_args$args;
				proxy_http_version 1.1;
				proxy_set_header Upgrade $http_upgrade;
				proxy_set_header Connection "upgrade";
				tcp_nodelay on;
			}`, jitsi.Name, jitsi.Namespace, jitsi.Spec.Domain)

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
