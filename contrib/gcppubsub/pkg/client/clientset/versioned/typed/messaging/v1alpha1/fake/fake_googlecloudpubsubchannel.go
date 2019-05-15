/*
Copyright 2019 The Knative Authors

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
	v1alpha1 "github.com/knative/eventing/contrib/gcppubsub/pkg/apis/messaging/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeGoogleCloudPubSubChannels implements GoogleCloudPubSubChannelInterface
type FakeGoogleCloudPubSubChannels struct {
	Fake *FakeMessagingV1alpha1
	ns   string
}

var googlecloudpubsubchannelsResource = schema.GroupVersionResource{Group: "messaging.knative.dev", Version: "v1alpha1", Resource: "googlecloudpubsubchannels"}

var googlecloudpubsubchannelsKind = schema.GroupVersionKind{Group: "messaging.knative.dev", Version: "v1alpha1", Kind: "GoogleCloudPubSubChannel"}

// Get takes name of the googleCloudPubSubChannel, and returns the corresponding googleCloudPubSubChannel object, and an error if there is any.
func (c *FakeGoogleCloudPubSubChannels) Get(name string, options v1.GetOptions) (result *v1alpha1.GoogleCloudPubSubChannel, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(googlecloudpubsubchannelsResource, c.ns, name), &v1alpha1.GoogleCloudPubSubChannel{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.GoogleCloudPubSubChannel), err
}

// List takes label and field selectors, and returns the list of GoogleCloudPubSubChannels that match those selectors.
func (c *FakeGoogleCloudPubSubChannels) List(opts v1.ListOptions) (result *v1alpha1.GoogleCloudPubSubChannelList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(googlecloudpubsubchannelsResource, googlecloudpubsubchannelsKind, c.ns, opts), &v1alpha1.GoogleCloudPubSubChannelList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.GoogleCloudPubSubChannelList{ListMeta: obj.(*v1alpha1.GoogleCloudPubSubChannelList).ListMeta}
	for _, item := range obj.(*v1alpha1.GoogleCloudPubSubChannelList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested googleCloudPubSubChannels.
func (c *FakeGoogleCloudPubSubChannels) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(googlecloudpubsubchannelsResource, c.ns, opts))

}

// Create takes the representation of a googleCloudPubSubChannel and creates it.  Returns the server's representation of the googleCloudPubSubChannel, and an error, if there is any.
func (c *FakeGoogleCloudPubSubChannels) Create(googleCloudPubSubChannel *v1alpha1.GoogleCloudPubSubChannel) (result *v1alpha1.GoogleCloudPubSubChannel, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(googlecloudpubsubchannelsResource, c.ns, googleCloudPubSubChannel), &v1alpha1.GoogleCloudPubSubChannel{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.GoogleCloudPubSubChannel), err
}

// Update takes the representation of a googleCloudPubSubChannel and updates it. Returns the server's representation of the googleCloudPubSubChannel, and an error, if there is any.
func (c *FakeGoogleCloudPubSubChannels) Update(googleCloudPubSubChannel *v1alpha1.GoogleCloudPubSubChannel) (result *v1alpha1.GoogleCloudPubSubChannel, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(googlecloudpubsubchannelsResource, c.ns, googleCloudPubSubChannel), &v1alpha1.GoogleCloudPubSubChannel{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.GoogleCloudPubSubChannel), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeGoogleCloudPubSubChannels) UpdateStatus(googleCloudPubSubChannel *v1alpha1.GoogleCloudPubSubChannel) (*v1alpha1.GoogleCloudPubSubChannel, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(googlecloudpubsubchannelsResource, "status", c.ns, googleCloudPubSubChannel), &v1alpha1.GoogleCloudPubSubChannel{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.GoogleCloudPubSubChannel), err
}

// Delete takes name of the googleCloudPubSubChannel and deletes it. Returns an error if one occurs.
func (c *FakeGoogleCloudPubSubChannels) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(googlecloudpubsubchannelsResource, c.ns, name), &v1alpha1.GoogleCloudPubSubChannel{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeGoogleCloudPubSubChannels) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(googlecloudpubsubchannelsResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &v1alpha1.GoogleCloudPubSubChannelList{})
	return err
}

// Patch applies the patch and returns the patched googleCloudPubSubChannel.
func (c *FakeGoogleCloudPubSubChannels) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.GoogleCloudPubSubChannel, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(googlecloudpubsubchannelsResource, c.ns, name, data, subresources...), &v1alpha1.GoogleCloudPubSubChannel{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.GoogleCloudPubSubChannel), err
}
