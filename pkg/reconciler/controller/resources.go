package controller

import (
	"github.com/aliok/websocket-channel/pkg/apis/channels/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/kmeta"
	"knative.dev/pkg/network"
)

const (
	PortName           = "http"
	PortNumber         = 80
	MessagingRoleLabel = "messaging.knative.dev/role"
	MessagingRole      = "websocket-channel"
)

// ServiceOption can be used to optionally modify the K8s service in CreateK8sService
type K8sServiceOption func(*corev1.Service) error

func createChannelServiceName(name string) string {
	return kmeta.ChildName(name, "-kn-channel")
}

// externalService is a functional option for CreateK8sService to create a K8s service of type ExternalName
// pointing to the specified service in a namespace.
func externalService(namespace, service string) K8sServiceOption {
	return func(svc *corev1.Service) error {
		svc.Spec = corev1.ServiceSpec{
			Type:         corev1.ServiceTypeExternalName,
			ExternalName: network.GetServiceHostname(service, namespace),
		}
		return nil
	}
}

// newK8sService creates a new Service for a Channel resource. It also sets the appropriate
// OwnerReferences on the resource so handleObject can discover the Channel resource that 'owns' it.
// As well as being garbage collected when the Channel is deleted.
func newK8sService(wsc *v1alpha1.WebSocketChannel, opts ...K8sServiceOption) (*corev1.Service, error) {
	// Add annotations
	svc := &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      createChannelServiceName(wsc.ObjectMeta.Name),
			Namespace: wsc.Namespace,
			Labels: map[string]string{
				MessagingRoleLabel: MessagingRole,
			},
			OwnerReferences: []metav1.OwnerReference{
				*kmeta.NewControllerRef(wsc),
			},
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name:     PortName,
					Protocol: corev1.ProtocolTCP,
					Port:     PortNumber,
				},
			},
		},
	}
	for _, opt := range opts {
		if err := opt(svc); err != nil {
			return nil, err
		}
	}
	return svc, nil
}