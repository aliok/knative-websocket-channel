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

// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	"context"

	v1alpha1 "github.com/aliok/websocket-channel/pkg/apis/channels/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeWebSocketChannels implements WebSocketChannelInterface
type FakeWebSocketChannels struct {
	Fake *FakeChannelsV1alpha1
	ns   string
}

var websocketchannelsResource = schema.GroupVersionResource{Group: "channels.aliok.github.com", Version: "v1alpha1", Resource: "websocketchannels"}

var websocketchannelsKind = schema.GroupVersionKind{Group: "channels.aliok.github.com", Version: "v1alpha1", Kind: "WebSocketChannel"}

// Get takes name of the webSocketChannel, and returns the corresponding webSocketChannel object, and an error if there is any.
func (c *FakeWebSocketChannels) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.WebSocketChannel, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(websocketchannelsResource, c.ns, name), &v1alpha1.WebSocketChannel{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.WebSocketChannel), err
}

// List takes label and field selectors, and returns the list of WebSocketChannels that match those selectors.
func (c *FakeWebSocketChannels) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.WebSocketChannelList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(websocketchannelsResource, websocketchannelsKind, c.ns, opts), &v1alpha1.WebSocketChannelList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.WebSocketChannelList{ListMeta: obj.(*v1alpha1.WebSocketChannelList).ListMeta}
	for _, item := range obj.(*v1alpha1.WebSocketChannelList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested webSocketChannels.
func (c *FakeWebSocketChannels) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(websocketchannelsResource, c.ns, opts))

}

// Create takes the representation of a webSocketChannel and creates it.  Returns the server's representation of the webSocketChannel, and an error, if there is any.
func (c *FakeWebSocketChannels) Create(ctx context.Context, webSocketChannel *v1alpha1.WebSocketChannel, opts v1.CreateOptions) (result *v1alpha1.WebSocketChannel, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(websocketchannelsResource, c.ns, webSocketChannel), &v1alpha1.WebSocketChannel{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.WebSocketChannel), err
}

// Update takes the representation of a webSocketChannel and updates it. Returns the server's representation of the webSocketChannel, and an error, if there is any.
func (c *FakeWebSocketChannels) Update(ctx context.Context, webSocketChannel *v1alpha1.WebSocketChannel, opts v1.UpdateOptions) (result *v1alpha1.WebSocketChannel, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(websocketchannelsResource, c.ns, webSocketChannel), &v1alpha1.WebSocketChannel{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.WebSocketChannel), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeWebSocketChannels) UpdateStatus(ctx context.Context, webSocketChannel *v1alpha1.WebSocketChannel, opts v1.UpdateOptions) (*v1alpha1.WebSocketChannel, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(websocketchannelsResource, "status", c.ns, webSocketChannel), &v1alpha1.WebSocketChannel{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.WebSocketChannel), err
}

// Delete takes name of the webSocketChannel and deletes it. Returns an error if one occurs.
func (c *FakeWebSocketChannels) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(websocketchannelsResource, c.ns, name), &v1alpha1.WebSocketChannel{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeWebSocketChannels) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(websocketchannelsResource, c.ns, listOpts)

	_, err := c.Fake.Invokes(action, &v1alpha1.WebSocketChannelList{})
	return err
}

// Patch applies the patch and returns the patched webSocketChannel.
func (c *FakeWebSocketChannels) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.WebSocketChannel, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(websocketchannelsResource, c.ns, name, pt, data, subresources...), &v1alpha1.WebSocketChannel{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.WebSocketChannel), err
}
