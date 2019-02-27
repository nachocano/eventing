/*
Copyright 2018 The Knative Authors

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
	"github.com/google/go-cmp/cmp"
	"github.com/knative/pkg/apis"
)

func (et *EventType) Validate() *apis.FieldError {
	return et.Spec.Validate().ViaField("spec")
}

func (et *EventTypeSpec) Validate() *apis.FieldError {
	var errs *apis.FieldError
	return errs
}

func (r *EventType) CheckImmutableFields(og apis.Immutable) *apis.FieldError {
	if og == nil {
		return nil
	}

	original, ok := og.(*EventType)
	if !ok {
		return &apis.FieldError{Message: "The provided original was not an EventType"}
	}

	// TODO check other fields.
	if diff := cmp.Diff(original.Spec.Origin, r.Spec.Origin); diff != "" {
		return &apis.FieldError{
			Message: "Immutable fields changed (-old +new)",
			Paths:   []string{"spec", "events"},
			Details: diff,
		}
	}
	return nil
}
