package controller

import (
	"context"

	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
)

// NewController initializes the controller and is called by the generated code.
// Registers event handlers to enqueue events.
func NewController(
	ctx context.Context,
	cmw configmap.Watcher,
) *controller.Impl {

}
