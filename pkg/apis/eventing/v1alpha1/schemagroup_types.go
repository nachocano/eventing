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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"
	"knative.dev/pkg/kmeta"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// SchemaGroup represents a collection of Schemas.
type SchemaGroup struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec defines the desired state of the SchemaGroup.
	Spec SchemaGroupSpec `json:"spec,omitempty"`

	// Status represents the current state of the SchemaGroup.
	// This data may be out of date.
	// +optional
	Status SchemaGroupStatus `json:"status,omitempty"`
}

var (
	// Check that SchemaGroup can be validated, can be defaulted, and has immutable fields.
	_ apis.Validatable = (*SchemaGroup)(nil)
	_ apis.Defaultable = (*SchemaGroup)(nil)

	// Check that SchemaGroup can return its spec untyped.
	_ apis.HasSpec = (*SchemaGroup)(nil)

	_ runtime.Object = (*SchemaGroup)(nil)

	// Check that we can create OwnerReferences to an SchemaGroup.
	_ kmeta.OwnerRefable = (*SchemaGroup)(nil)

	// Check that the type conforms to the duck Knative Resource shape.
	//_ duckv1.KRShaped = (*SchemaGroup)(nil)
)

type SchemaGroupSpec struct {
	// +optional
	Format string `json:"format,omitempty"`
	// +optional
	Description string                        `json:"description,omitempty"`
	Schemas     []corev1.LocalObjectReference `json:"schemas,omitempty"`
}

// SchemaGroupStatus represents the current state of a SchemaGroup.
type SchemaGroupStatus struct {
	// inherits duck/v1 Status, which currently provides:
	// * ObservedGeneration - the 'Generation' of the Service that was last processed by the controller.
	// * Conditions - the latest available observations of a resource's current state.
	duckv1.Status `json:",inline"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// SchemaGroupList is a collection of SchemaGroups.
type SchemaGroupList struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SchemaGroup `json:"items"`
}

// GetGroupVersionKind returns GroupVersionKind for SchemaGroup
func (p *SchemaGroup) GetGroupVersionKind() schema.GroupVersionKind {
	return SchemeGroupVersion.WithKind("SchemaGroup")
}

// GetUntypedSpec returns the spec of the SchemaGroup.
func (e *SchemaGroup) GetUntypedSpec() interface{} {
	return e.Spec
}

// GetStatus retrieves the status of the SchemaGroup. Implements the KRShaped interface.
func (t *SchemaGroup) GetStatus() *duckv1.Status {
	return &t.Status.Status
}
