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

package v1

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"knative.dev/pkg/apis"
	"knative.dev/pkg/apis/duck/ducktypes"
)

type EventTypeable struct {
	// TODO create a struct instead of just a list of strings.
	EventTypes []string `json:"eventTypes,omitempty"`
}

// +genduck
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type EventTypeableType struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Status EventTypeableStatus `json:"status,omitempty"`
}

type EventTypeableStatus struct {
	// TODO add path  merge strategy
	EventTypeable `json:",inline"`
}

// Verify EventTypeableType resources meet duck contracts.
var (
	_ apis.Listable         = (*EventTypeableType)(nil)
	_ ducktypes.Populatable = (*EventTypeableType)(nil)
)

// GetFullType implements duck.Implementable
func (*EventTypeableType) GetFullType() ducktypes.Populatable {
	return &EventTypeableType{}
}

// ConvertTo implements apis.Convertible
func (a *EventTypeableType) ConvertTo(ctx context.Context, to apis.Convertible) error {
	return fmt.Errorf("v1 is the highest known version, got: %T", to)
}

// ConvertFrom implements apis.Convertible
func (a *EventTypeableType) ConvertFrom(ctx context.Context, from apis.Convertible) error {
	return fmt.Errorf("v1 is the highest known version, got: %T", from)
}

// Populate implements duck.Populatable
func (t *EventTypeableType) Populate() {
	t.Status = EventTypeableStatus{
		EventTypeable: EventTypeable{
			EventTypes: []string{
				"mytype",
				"myothertype",
			},
		},
	}
}

// GetListType implements apis.Listable
func (*EventTypeableType) GetListType() runtime.Object {
	return &EventTypeableTypeList{}
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// EventTypeableTypeList is a list of EventTypeableType resources
type EventTypeableTypeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []EventTypeableTypeList `json:"items"`
}
