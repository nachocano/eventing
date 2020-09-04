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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"
	"knative.dev/pkg/kmeta"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type Schema struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec defines the desired state of the Schema.
	Spec SchemaSpec `json:"spec,omitempty"`

	// Status represents the current state of the Schema.
	// This data may be out of date.
	// +optional
	Status SchemaStatus `json:"status,omitempty"`
}

var (
	// Check that Schema can be validated, can be defaulted, and has immutable fields.
	_ apis.Validatable = (*Schema)(nil)
	_ apis.Defaultable = (*Schema)(nil)

	// Check that Schema can return its spec untyped.
	_ apis.HasSpec = (*Schema)(nil)

	_ runtime.Object = (*Schema)(nil)

	// Check that we can create OwnerReferences to an Schema.
	_ kmeta.OwnerRefable = (*Schema)(nil)

	// Check that the type conforms to the duck Knative Resource shape.
	//_ duckv1.KRShaped = (*Schema)(nil)
)

type SchemaSpec struct {
	// +optional
	Format string `json:"format"`
	// +optional
	Description string `json:"description,omitempty"`

	Versions []Version `json:"versions"`
}

type Version struct {
	Version int32 `json:"version"`
	// +optional
	Description string `json:"description,omitempty"`
	Body        string `json:"body"`
}

// SchemaStatus represents the current state of a Schema.
type SchemaStatus struct {
	// inherits duck/v1 Status, which currently provides:
	// * ObservedGeneration - the 'Generation' of the Service that was last processed by the controller.
	// * Conditions - the latest available observations of a resource's current state.
	duckv1.Status `json:",inline"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// SchemaList is a collection of Schemas.
type SchemaList struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Schema `json:"items"`
}

// GetGroupVersionKind returns GroupVersionKind for Schema
func (p *Schema) GetGroupVersionKind() schema.GroupVersionKind {
	return SchemeGroupVersion.WithKind("Schema")
}

// GetUntypedSpec returns the spec of the Schema.
func (e *Schema) GetUntypedSpec() interface{} {
	return e.Spec
}

// GetStatus retrieves the status of the Schema. Implements the KRShaped interface.
func (t *Schema) GetStatus() *duckv1.Status {
	return &t.Status.Status
}
