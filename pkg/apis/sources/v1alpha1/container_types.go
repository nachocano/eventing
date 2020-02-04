/*
Copyright 2020 The Knative Authors

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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"knative.dev/pkg/apis"
	"knative.dev/pkg/kmeta"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:defaulter-gen=true

type ContainerSource struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ContainerSourceSpec   `json:"spec"`
	Status ContainerSourceStatus `json:"status"`
}

// Check the interfaces that SinkBinding should be implementing.
var (
	_ runtime.Object     = (*ContainerSource)(nil)
	_ kmeta.OwnerRefable = (*ContainerSource)(nil)
	_ apis.Validatable   = (*ContainerSource)(nil)
	_ apis.Defaultable   = (*ContainerSource)(nil)
	_ apis.HasSpec       = (*ContainerSource)(nil)
)

// SinkBindingSpec holds the desired state of the SinkBinding (from the client).
type ContainerSourceSpec struct {
	SinkBindingSpec `json:",inline"`
}

const (
	// SinkBindingConditionReady is configured to indicate whether the Binding
	// has been configured for resources subject to its runtime contract.
	ContainerSourceConditionReady = apis.ConditionReady
)

// SinkBindingStatus communicates the observed state of the SinkBinding (from the controller).
type ContainerSourceStatus struct {
	SinkBindingStatus `json:",inline"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// SinkBindingList contains a list of SinkBinding
type ContainerSourceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ContainerSource `json:"items"`
}
