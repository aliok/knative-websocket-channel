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

// Code generated by lister-gen. DO NOT EDIT.

package v1alpha1

import (
	v1alpha1 "github.com/aliok/websocket-channel/pkg/apis/channels/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// WebSocketChannelLister helps list WebSocketChannels.
// All objects returned here must be treated as read-only.
type WebSocketChannelLister interface {
	// List lists all WebSocketChannels in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha1.WebSocketChannel, err error)
	// WebSocketChannels returns an object that can list and get WebSocketChannels.
	WebSocketChannels(namespace string) WebSocketChannelNamespaceLister
	WebSocketChannelListerExpansion
}

// webSocketChannelLister implements the WebSocketChannelLister interface.
type webSocketChannelLister struct {
	indexer cache.Indexer
}

// NewWebSocketChannelLister returns a new WebSocketChannelLister.
func NewWebSocketChannelLister(indexer cache.Indexer) WebSocketChannelLister {
	return &webSocketChannelLister{indexer: indexer}
}

// List lists all WebSocketChannels in the indexer.
func (s *webSocketChannelLister) List(selector labels.Selector) (ret []*v1alpha1.WebSocketChannel, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.WebSocketChannel))
	})
	return ret, err
}

// WebSocketChannels returns an object that can list and get WebSocketChannels.
func (s *webSocketChannelLister) WebSocketChannels(namespace string) WebSocketChannelNamespaceLister {
	return webSocketChannelNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// WebSocketChannelNamespaceLister helps list and get WebSocketChannels.
// All objects returned here must be treated as read-only.
type WebSocketChannelNamespaceLister interface {
	// List lists all WebSocketChannels in the indexer for a given namespace.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha1.WebSocketChannel, err error)
	// Get retrieves the WebSocketChannel from the indexer for a given namespace and name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*v1alpha1.WebSocketChannel, error)
	WebSocketChannelNamespaceListerExpansion
}

// webSocketChannelNamespaceLister implements the WebSocketChannelNamespaceLister
// interface.
type webSocketChannelNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all WebSocketChannels in the indexer for a given namespace.
func (s webSocketChannelNamespaceLister) List(selector labels.Selector) (ret []*v1alpha1.WebSocketChannel, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.WebSocketChannel))
	})
	return ret, err
}

// Get retrieves the WebSocketChannel from the indexer for a given namespace and name.
func (s webSocketChannelNamespaceLister) Get(name string) (*v1alpha1.WebSocketChannel, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1alpha1.Resource("websocketchannel"), name)
	}
	return obj.(*v1alpha1.WebSocketChannel), nil
}
