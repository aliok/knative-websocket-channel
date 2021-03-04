/*
Copyright 2021 The Knative Authors

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

package controller

import (
	"context"

	"github.com/aliok/websocket-channel/pkg/client/injection/informers/channels/v1alpha1/websocketchannel"

	"github.com/kelseyhightower/envconfig"
	"k8s.io/client-go/tools/cache"
	kubeclient "knative.dev/pkg/client/injection/kube/client"

	"knative.dev/pkg/client/injection/kube/informers/apps/v1/deployment"
	"knative.dev/pkg/client/injection/kube/informers/core/v1/endpoints"
	"knative.dev/pkg/client/injection/kube/informers/core/v1/service"
	"knative.dev/pkg/client/injection/kube/informers/core/v1/serviceaccount"
	"knative.dev/pkg/client/injection/kube/informers/rbac/v1/rolebinding"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/logging"
	"knative.dev/pkg/system"

	websocketchannelreconciler "github.com/aliok/websocket-channel/pkg/client/injection/reconciler/channels/v1alpha1/websocketchannel"
)

const dispatcherName = "websocket-ch-dispatcher"

type envConfig struct {
	Image string `envconfig:"DISPATCHER_IMAGE" required:"true"`
}

// NewController initializes the controller and is called by the generated code.
// Registers event handlers to enqueue events.
func NewController(
	ctx context.Context,
	cmw configmap.Watcher,
) *controller.Impl {
	logger := logging.FromContext(ctx)
	websocketchannelInformer := websocketchannel.Get(ctx)
	deploymentInformer := deployment.Get(ctx)
	serviceInformer := service.Get(ctx)
	endpointsInformer := endpoints.Get(ctx)
	serviceAccountInformer := serviceaccount.Get(ctx)
	roleBindingInformer := rolebinding.Get(ctx)

	r := &Reconciler{
		kubeClientSet:            kubeclient.Get(ctx),
		systemNamespace:          system.Namespace(),
		websocketchannelLister:   websocketchannelInformer.Lister(),
		websocketchannelInformer: websocketchannelInformer.Informer(),
		deploymentLister:         deploymentInformer.Lister(),
		serviceLister:            serviceInformer.Lister(),
		endpointsLister:          endpointsInformer.Lister(),
		serviceAccountLister:     serviceAccountInformer.Lister(),
		roleBindingLister:        roleBindingInformer.Lister(),
	}

	env := &envConfig{}
	if err := envconfig.Process("", env); err != nil {
		logger.Panicf("unable to process web socket channel's required environment variables: %v", err)
	}

	if env.Image == "" {
		logger.Panic("unable to process web socket channel's required environment variables (missing DISPATCHER_IMAGE)")
	}

	r.dispatcherImage = env.Image

	impl := websocketchannelreconciler.NewImpl(ctx, r)

	logger.Info("Setting up event handlers")
	websocketchannelInformer.Informer().AddEventHandler(controller.HandleAll(impl.Enqueue))

	// Set up watches for dispatcher resources we care about, since any changes to these
	// resources will affect our Channels. So, set up a watch here, that will cause
	// a global Resync for all the channels to take stock of their health when these change.

	// Call GlobalResync on websocketchannels.
	grCh := func(obj interface{}) {
		impl.GlobalResync(websocketchannelInformer.Informer())
	}

	// watch deployments, services, endpoints, etc. with name `websocker-dispatcher`
	// the names are hardcoded in YAML files
	deploymentInformer.Informer().AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: controller.FilterWithName(dispatcherName),
		Handler:    controller.HandleAll(grCh),
	})
	serviceInformer.Informer().AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: controller.FilterWithName(dispatcherName),
		Handler:    controller.HandleAll(grCh),
	})
	endpointsInformer.Informer().AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: controller.FilterWithName(dispatcherName),
		Handler:    controller.HandleAll(grCh),
	})
	serviceAccountInformer.Informer().AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: controller.FilterWithName(dispatcherName),
		Handler:    controller.HandleAll(grCh),
	})
	roleBindingInformer.Informer().AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: controller.FilterWithName(dispatcherName),
		Handler:    controller.HandleAll(grCh),
	})

	return impl
}
