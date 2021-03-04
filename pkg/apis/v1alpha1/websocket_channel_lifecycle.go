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

package v1alpha1

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"knative.dev/pkg/apis"
	v1 "knative.dev/pkg/apis/duck/v1"
)

var wscCondSet = apis.NewLivingConditionSet(
	WebsocketChannelConditionDispatcherReady,
	WebsocketChannelConditionServiceReady,
	WebsocketChannelConditionEndpointsReady,
	WebsocketChannelConditionAddressable,
	WebsocketChannelConditionChannelServiceReady,
)

const (
	// WebsocketChannelConditionReady has status True when all subconditions below have been set to True.
	WebsocketChannelConditionReady = apis.ConditionReady

	// WebsocketChannelConditionDispatcherReady has status True when a Dispatcher deployment is ready
	// Keyed off appsv1.DeploymentAvailable, which means minimum available replicas required are up
	// and running for at least minReadySeconds.
	WebsocketChannelConditionDispatcherReady apis.ConditionType = "DispatcherReady"

	// WebsocketChannelConditionServiceReady has status True when a k8s Service is ready. This
	// basically just means it exists because there's no meaningful status in Service. See Endpoints
	// below.
	WebsocketChannelConditionServiceReady apis.ConditionType = "ServiceReady"

	// WebsocketChannelConditionEndpointsReady has status True when a k8s Service Endpoints are backed
	// by at least one endpoint.
	WebsocketChannelConditionEndpointsReady apis.ConditionType = "EndpointsReady"

	// WebsocketChannelConditionAddressable has status true when this WebSocketChannel meets
	// the Addressable contract and has a non-empty hostname.
	WebsocketChannelConditionAddressable apis.ConditionType = "Addressable"

	// WebsocketChannelConditionServiceReady has status True when a k8s Service representing the channel is ready.
	// Because this uses ExternalName, there are no endpoints to check.
	WebsocketChannelConditionChannelServiceReady apis.ConditionType = "ChannelServiceReady"
)

// GetConditionSet retrieves the condition set for this resource. Implements the KRShaped interface.
func (*WebSocketChannel) GetConditionSet() apis.ConditionSet {
	return wscCondSet
}

// GetGroupVersionKind returns GroupVersionKind for WebsocketChannels
func (*WebSocketChannel) GetGroupVersionKind() schema.GroupVersionKind {
	return SchemeGroupVersion.WithKind("WebSocketChannel")
}

// GetUntypedSpec returns the spec of the WebSocketChannel.
func (i *WebSocketChannel) GetUntypedSpec() interface{} {
	return i.Spec
}

// GetCondition returns the condition currently associated with the given type, or nil.
func (wscs *WebSocketChannelStatus) GetCondition(t apis.ConditionType) *apis.Condition {
	return wscCondSet.Manage(wscs).GetCondition(t)
}

// IsReady returns true if the resource is ready overall.
func (wscs *WebSocketChannelStatus) IsReady() bool {
	return wscCondSet.Manage(wscs).IsHappy()
}

// InitializeConditions sets relevant unset conditions to Unknown state.
func (wscs *WebSocketChannelStatus) InitializeConditions() {
	wscCondSet.Manage(wscs).InitializeConditions()
}

func (wscs *WebSocketChannelStatus) SetAddress(url *apis.URL) {
	wscs.Address = &v1.Addressable{URL: url}
	if url != nil {
		wscCondSet.Manage(wscs).MarkTrue(WebsocketChannelConditionAddressable)
	} else {
		wscCondSet.Manage(wscs).MarkFalse(WebsocketChannelConditionAddressable, "emptyHostname", "hostname is the empty string")
	}
}

func (wscs *WebSocketChannelStatus) MarkDispatcherFailed(reason, messageFormat string, messageA ...interface{}) {
	wscCondSet.Manage(wscs).MarkFalse(WebsocketChannelConditionDispatcherReady, reason, messageFormat, messageA...)
}

func (wscs *WebSocketChannelStatus) MarkDispatcherUnknown(reason, messageFormat string, messageA ...interface{}) {
	wscCondSet.Manage(wscs).MarkUnknown(WebsocketChannelConditionDispatcherReady, reason, messageFormat, messageA...)
}

// TODO: Unify this with the ones from Eventing. Say: Broker, Trigger.
func (wscs *WebSocketChannelStatus) PropagateDispatcherStatus(ds *appsv1.DeploymentStatus) {
	for _, cond := range ds.Conditions {
		if cond.Type == appsv1.DeploymentAvailable {
			if cond.Status == corev1.ConditionTrue {
				wscCondSet.Manage(wscs).MarkTrue(WebsocketChannelConditionDispatcherReady)
			} else if cond.Status == corev1.ConditionFalse {
				wscs.MarkDispatcherFailed("DispatcherDeploymentFalse", "The status of Dispatcher Deployment is False: %s : %s", cond.Reason, cond.Message)
			} else if cond.Status == corev1.ConditionUnknown {
				wscs.MarkDispatcherUnknown("DispatcherDeploymentUnknown", "The status of Dispatcher Deployment is Unknown: %s : %s", cond.Reason, cond.Message)
			}
		}
	}
}

func (wscs *WebSocketChannelStatus) MarkServiceFailed(reason, messageFormat string, messageA ...interface{}) {
	wscCondSet.Manage(wscs).MarkFalse(WebsocketChannelConditionServiceReady, reason, messageFormat, messageA...)
}

func (wscs *WebSocketChannelStatus) MarkServiceUnknown(reason, messageFormat string, messageA ...interface{}) {
	wscCondSet.Manage(wscs).MarkUnknown(WebsocketChannelConditionServiceReady, reason, messageFormat, messageA...)
}

func (wscs *WebSocketChannelStatus) MarkServiceTrue() {
	wscCondSet.Manage(wscs).MarkTrue(WebsocketChannelConditionServiceReady)
}

func (wscs *WebSocketChannelStatus) MarkChannelServiceFailed(reason, messageFormat string, messageA ...interface{}) {
	wscCondSet.Manage(wscs).MarkFalse(WebsocketChannelConditionChannelServiceReady, reason, messageFormat, messageA...)
}

func (wscs *WebSocketChannelStatus) MarkChannelServiceUnknown(reason, messageFormat string, messageA ...interface{}) {
	wscCondSet.Manage(wscs).MarkUnknown(WebsocketChannelConditionChannelServiceReady, reason, messageFormat, messageA...)
}

func (wscs *WebSocketChannelStatus) MarkChannelServiceTrue() {
	wscCondSet.Manage(wscs).MarkTrue(WebsocketChannelConditionChannelServiceReady)
}

func (wscs *WebSocketChannelStatus) MarkEndpointsFailed(reason, messageFormat string, messageA ...interface{}) {
	wscCondSet.Manage(wscs).MarkFalse(WebsocketChannelConditionEndpointsReady, reason, messageFormat, messageA...)
}

func (wscs *WebSocketChannelStatus) MarkEndpointsUnknown(reason, messageFormat string, messageA ...interface{}) {
	wscCondSet.Manage(wscs).MarkUnknown(WebsocketChannelConditionEndpointsReady, reason, messageFormat, messageA...)
}

func (wscs *WebSocketChannelStatus) MarkEndpointsTrue() {
	wscCondSet.Manage(wscs).MarkTrue(WebsocketChannelConditionEndpointsReady)
}
