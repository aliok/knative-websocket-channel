package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	eventingduckv1 "knative.dev/eventing/pkg/apis/duck/v1"
	"knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"
	"knative.dev/pkg/kmeta"
)

// +genclient
// +genreconciler
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type WebSocketChannel struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec defines the desired state of the Channel.
	Spec WebSocketChannelSpec `json:"spec,omitempty"`

	// Status represents the current state of the Channel. This data may be out of
	// date.
	// +optional
	Status WebSocketChannelStatus `json:"status,omitempty"`
}

var (
	// Check that WebSocketChannel can be validated and defaulted.
	_ apis.Validatable = (*WebSocketChannel)(nil)
	_ apis.Defaultable = (*WebSocketChannel)(nil)

	// Check that WebSocketChannel can return its spec untyped.
	_ apis.HasSpec = (*WebSocketChannel)(nil)

	_ runtime.Object = (*WebSocketChannel)(nil)

	// Check that we can create OwnerReferences to an WebSocketChannel.
	_ kmeta.OwnerRefable = (*WebSocketChannel)(nil)

	// Check that the type conforms to the duck Knative Resource shape.
	_ duckv1.KRShaped = (*WebSocketChannel)(nil)
)

// WebSocketChannelSpec defines which subscribers have expressed interest in
// receiving events from this WebSocketChannel.
// arguments for a Channel.
type WebSocketChannelSpec struct {
	// Channel conforms to Duck type Channelable.
	eventingduckv1.ChannelableSpec `json:",inline"`
}

// ChannelStatus represents the current state of a Channel.
type WebSocketChannelStatus struct {
	// Channel conforms to Duck type Channelable.
	eventingduckv1.ChannelableStatus `json:",inline"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// WebSocketChannelList is a collection of WebsocketChannels.
type WebSocketChannelList struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []WebSocketChannel `json:"items"`
}

// GetStatus retrieves the status of the WebSocketChannel. Implements the KRShaped interface.
func (wsc *WebSocketChannel) GetStatus() *duckv1.Status {
	return &wsc.Status.Status
}
