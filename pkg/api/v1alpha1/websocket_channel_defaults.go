package v1alpha1

import (
	"context"

	"knative.dev/eventing/pkg/apis/messaging"
)

func (wsc *WebSocketChannel) SetDefaults(ctx context.Context) {
	// Set the duck subscription to the stored version of the duck
	// we support. Reason for this is that the stored version will
	// not get a chance to get modified, but for newer versions
	// conversion webhook will be able to take a crack at it and
	// can modify it to match the duck shape.
	if wsc.Annotations == nil {
		wsc.Annotations = make(map[string]string)
	}
	if _, ok := wsc.Annotations[messaging.SubscribableDuckVersionAnnotation]; !ok {
		wsc.Annotations[messaging.SubscribableDuckVersionAnnotation] = "v1"
	}

	wsc.Spec.SetDefaults(ctx)
}

func (imcs *WebSocketChannelSpec) SetDefaults(_ context.Context) {
	// TODO: Nothing to default here...
}
