// +build !ignore_autogenerated

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

// Code generated by deepcopy-gen. DO NOT EDIT.

package v1alpha1

import (
	duckv1alpha1 "github.com/knative/eventing/pkg/apis/duck/v1alpha1"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NatssChannel) DeepCopyInto(out *NatssChannel) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NatssChannel.
func (in *NatssChannel) DeepCopy() *NatssChannel {
	if in == nil {
		return nil
	}
	out := new(NatssChannel)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *NatssChannel) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NatssChannelList) DeepCopyInto(out *NatssChannelList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	out.ListMeta = in.ListMeta
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]NatssChannel, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NatssChannelList.
func (in *NatssChannelList) DeepCopy() *NatssChannelList {
	if in == nil {
		return nil
	}
	out := new(NatssChannelList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *NatssChannelList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NatssChannelSpec) DeepCopyInto(out *NatssChannelSpec) {
	*out = *in
	if in.Subscribable != nil {
		in, out := &in.Subscribable, &out.Subscribable
		*out = new(duckv1alpha1.Subscribable)
		(*in).DeepCopyInto(*out)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NatssChannelSpec.
func (in *NatssChannelSpec) DeepCopy() *NatssChannelSpec {
	if in == nil {
		return nil
	}
	out := new(NatssChannelSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NatssChannelStatus) DeepCopyInto(out *NatssChannelStatus) {
	*out = *in
	in.Status.DeepCopyInto(&out.Status)
	in.Address.DeepCopyInto(&out.Address)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NatssChannelStatus.
func (in *NatssChannelStatus) DeepCopy() *NatssChannelStatus {
	if in == nil {
		return nil
	}
	out := new(NatssChannelStatus)
	in.DeepCopyInto(out)
	return out
}
