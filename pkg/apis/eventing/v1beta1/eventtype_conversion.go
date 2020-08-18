/*
Copyright 2020 The Knative Authors.

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

package v1beta1

import (
	"context"
	"fmt"

	"knative.dev/eventing/pkg/apis/eventing/v1"
	"knative.dev/pkg/apis"
)

// ConvertTo implements apis.Convertible
func (source *EventType) ConvertTo(ctx context.Context, to apis.Convertible) error {
	switch sink := to.(type) {
	case *v1.EventType:
		sink.ObjectMeta = source.ObjectMeta
		sink.Spec.Type = source.Spec.Type
		sink.Spec.Schema = source.Spec.Schema
		sink.Spec.SchemaData = source.Spec.SchemaData
		sink.Spec.Description = source.Spec.Description
		return nil
	default:
		return fmt.Errorf("unknown version, got: %T", sink)
	}
}

// ConvertFrom implements apis.Convertible
func (sink *EventType) ConvertFrom(ctx context.Context, from apis.Convertible) error {
	switch source := from.(type) {
	case *v1.EventType:
		sink.ObjectMeta = source.ObjectMeta
		sink.Spec.Type = source.Spec.Type
		sink.Spec.Schema = source.Spec.Schema
		sink.Spec.SchemaData = source.Spec.SchemaData
		sink.Spec.Description = source.Spec.Description
		return nil
	default:
		return fmt.Errorf("unknown version, got: %T", source)
	}
}
