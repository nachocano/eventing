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

package resources

import (
	"crypto/md5"
	"fmt"

	"knative.dev/eventing/pkg/apis/eventing/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	duckv1alpha1 "knative.dev/eventing/pkg/apis/duck/v1alpha1"
	"knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"
	"knative.dev/pkg/ptr"
)

// EventTypeArgs are the arguments needed to create an EventType for a Source.
type EventTypeArgs struct {
	Source         *duckv1.Source
	CrdName        string
	CeType         string
	CeSchema       *apis.URL
	SourceTemplate string
	Extensions     []v1.Extension
	Description    string
}

func MakeEventType(args *EventTypeArgs) *v1.EventType {
	// Name it with the hash of the concatenation of the three fields.
	// Cannot generate a fixed name based on type+UUID, because long type names might be cut, and we end up trying to create
	// event types with the same name.
	fixedName := fmt.Sprintf("%x", md5.Sum([]byte(args.CeType+args.CeSchema.String()+args.CrdName)))
	return &v1.EventType{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fixedName,
			Labels:    Labels(args.CrdName),
			Namespace: args.Source.Namespace,
			OwnerReferences: []metav1.OwnerReference{{
				APIVersion:         args.Source.APIVersion,
				Kind:               args.Source.Kind,
				Name:               args.Source.Name,
				UID:                args.Source.UID,
				BlockOwnerDeletion: ptr.Bool(true),
				Controller:         ptr.Bool(true),
			}},
		},
		Spec: v1.EventTypeSpec{
			Type:           args.CeType,
			SourceTemplate: args.SourceTemplate,
			Extensions:     args.Extensions,
			Description:    args.Description,
			// TODO won't be able to set a ref to the schema hosted here.
			Schema: &duckv1alpha1.Schema{
				URI: args.CeSchema,
			},
			// TODO add more stuff here.
		},
	}
}
