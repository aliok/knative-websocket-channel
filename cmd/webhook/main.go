/*
Copyright 2018 The Knative Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"context"

	channelsv1alpha1 "github.com/aliok/websocket-channel/pkg/apis/channels/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/injection/sharedmain"
	"knative.dev/pkg/signals"
	"knative.dev/pkg/webhook"
	"knative.dev/pkg/webhook/certificates"
	"knative.dev/pkg/webhook/resourcesemantics"
	"knative.dev/pkg/webhook/resourcesemantics/defaulting"
	"knative.dev/pkg/webhook/resourcesemantics/validation"
)

var ourTypes = map[schema.GroupVersionKind]resourcesemantics.GenericCRD{
	// For group channels.aliok.github.com
	// v1alpha1
	channelsv1alpha1.SchemeGroupVersion.WithKind("WebSocketChannel"): &channelsv1alpha1.WebSocketChannel{},
}

var callbacks = map[schema.GroupVersionKind]validation.Callback{}

func NewDefaultingAdmissionController(ctx context.Context, _ configmap.Watcher) *controller.Impl {

	return defaulting.NewAdmissionController(ctx,

		// Name of the resource webhook.
		"defaults.webhook.channels.aliok.github.com",

		// The path on which to serve the webhook.
		"/defaulting",

		// The resources to default.
		ourTypes,

		// A function that infuses the context passed to Validate/SetDefaults with custom metadata.
		nil,

		// Whether to disallow unknown fields.
		true,
	)
}

func NewValidationAdmissionController(ctx context.Context, cmw configmap.Watcher) *controller.Impl {

	return validation.NewAdmissionController(ctx,

		// Name of the resource webhook.
		"validation.webhook.channels.aliok.github.com",

		// The path on which to serve the webhook.
		"/resource-validation",

		// The resources to validate.
		ourTypes,

		// A function that infuses the context passed to Validate/SetDefaults with custom metadata.
		nil,

		// Whether to disallow unknown fields.
		true,

		// Extra validating callbacks to be applied to resources.
		callbacks,
	)
}

func main() {
	// Set up a signal context with our webhook options
	ctx := webhook.WithOptions(signals.NewContext(), webhook.Options{
		ServiceName: webhook.NameFromEnv(),
		Port:        webhook.PortFromEnv(8443),
		// SecretName must match the name of the Secret created in the configuration.
		SecretName: "websocket-channel-webhook-certs",
	})

	sharedmain.WebhookMainWithContext(ctx, webhook.NameFromEnv(),
		certificates.NewController,
		NewValidationAdmissionController,
		NewDefaultingAdmissionController,
	)
}
