/*
 * Copyright 2020 The Knative Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	duckv1alpha1 "knative.dev/eventing/pkg/apis/duck/v1alpha1"
	"knative.dev/pkg/apis"
	"knative.dev/pkg/kmeta"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type EventType struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec defines the desired state of the EventType.
	Spec EventTypeSpec `json:"spec,omitempty"`
}

var (
	// Check that EventType can be validated, can be defaulted, and has immutable fields.
	_ apis.Validatable = (*EventType)(nil)
	_ apis.Defaultable = (*EventType)(nil)

	// Check that EventType can return its spec untyped.
	_ apis.HasSpec = (*EventType)(nil)

	_ runtime.Object = (*EventType)(nil)

	// Check that we can create OwnerReferences to an EventType.
	_ kmeta.OwnerRefable = (*EventType)(nil)
)

type EventTypeSpec struct {
	// Type represents the CloudEvents type. It is authoritative.
	Type string `json:"type"`

	Schema *duckv1alpha1.Schema `json:"schema,omitempty"`

	// ContentType is the ContentType of the data payload.
	// TODO should be part of Schema?
	ContentType string `json:"contentType,omitempty"`

	// Description is an optional field used to describe the EventType, in any meaningful way.
	// +optional
	Description string `json:"description,omitempty"`

	// SourceTemplate is an optional field used to describe the CloudEvents source template.
	// If provided it should be a URI Template according to https://tools.ietf.org/html/rfc6570
	// +optional
	SourceTemplate string `json:"sourceTemplate,omitempty"`

	// Extensions is a list of CloudEvents extension attributes.
	Extensions []Extension `json:"extensions,omitempty"`
}

type Extension struct {
	// Name is the name of the extension. It MUST adhere to the CloudEvents context attrbibute naming rules.
	Name string `json:"name"`
	// Type refers to the data type of the extension attribute. It MUST adhere to the CloudEvents type system https://github.com/cloudevents/spec/blob/master/spec.md#type-system.
	Type string `json:"type"`
	// SpecURL is an attribute pointing to the specification that defines the extension.
	SpecURL string `json:"specUrl,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// EventTypeList is a collection of EventTypes.
type EventTypeList struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []EventType `json:"items"`
}

// GetGroupVersionKind returns GroupVersionKind for EventType
func (p *EventType) GetGroupVersionKind() schema.GroupVersionKind {
	return SchemeGroupVersion.WithKind("EventType")
}

// GetUntypedSpec returns the spec of the EventType.
func (e *EventType) GetUntypedSpec() interface{} {
	return e.Spec
}
