package v1alpha1

import (
	"context"
	"fmt"

	"knative.dev/pkg/apis"
)

func (wsc *WebSocketChannel) Validate(ctx context.Context) *apis.FieldError {
	errs := wsc.Spec.Validate(ctx).ViaField("spec")

	return errs
}

func (wsc *WebSocketChannelSpec) Validate(_ context.Context) *apis.FieldError {
	var errs *apis.FieldError
	for i, subscriber := range wsc.SubscribableSpec.Subscribers {
		if subscriber.ReplyURI == nil && subscriber.SubscriberURI == nil {
			fe := apis.ErrMissingField("replyURI", "subscriberURI")
			fe.Details = "expected at least one of, got none"
			errs = errs.Also(fe.ViaField(fmt.Sprintf("subscriber[%d]", i)).ViaField("subscribable"))
		}
	}

	return errs
}
