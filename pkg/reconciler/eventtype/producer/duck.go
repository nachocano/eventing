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

package producer

import (
	"context"
	"encoding/json"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime/schema"

	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
	duckv1 "knative.dev/eventing/pkg/apis/duck/v1"
	"knative.dev/eventing/pkg/apis/eventing"
	"knative.dev/eventing/pkg/apis/eventing/v1beta1"
	clientset "knative.dev/eventing/pkg/client/clientset/versioned"
	listers "knative.dev/eventing/pkg/client/listers/eventing/v1beta1"
	"knative.dev/eventing/pkg/reconciler/eventtype/producer/resources"
	"knative.dev/pkg/apis"
	"knative.dev/pkg/logging"
)

type Reconciler struct {
	// eventingClientSet allows us to configure Eventing objects
	eventingClientSet clientset.Interface

	// listers index properties about resources
	eventTypeLister listers.EventTypeLister
	resourceLister  cache.GenericLister

	gvr schema.GroupVersionResource
}

// eventTypeEntry refers to an entry in the registry.knative.dev/eventTypes annotation.
type eventTypeEntry struct {
	Type        string `json:"type"`
	Schema      string `json:"schema,omitempty"`
	Description string `json:"description,omitempty"`
}

func (r *Reconciler) Reconcile(ctx context.Context, key string) error {
	// Convert the namespace/name string into a distinct namespace and name
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		logging.FromContext(ctx).Error("invalid resource key")
		return nil
	}

	// Get the Resource with this namespace/name
	runtimeObj, err := r.resourceLister.ByNamespace(namespace).Get(name)

	var ok bool
	var original *duckv1.Resource
	if original, ok = runtimeObj.(*duckv1.Resource); !ok {
		logging.FromContext(ctx).Errorw("runtime object is not convertible to Resource duck type: ", zap.Any("runtimeObj", runtimeObj))
		// Avoid re-enqueuing.
		return nil
	}

	if apierrs.IsNotFound(err) {
		// The resource may no longer exist, in which case we stop processing.
		logging.FromContext(ctx).Error("Resource in work queue no longer exists")
		return nil
	} else if err != nil {
		return err
	}

	// Don't modify the informers copy
	orig := original.DeepCopy()
	// Reconcile this copy of the Resource. We do not control the Resource, so do not update status.
	return r.reconcile(ctx, orig)
}

func (r *Reconciler) reconcile(ctx context.Context, resource *duckv1.Resource) error {
	// Reconcile the eventTypes for this resource.
	err := r.reconcileEventTypes(ctx, resource)
	if err != nil {
		logging.FromContext(ctx).Error("Error reconciling event types for Resource")
		return err
	}
	return nil
}

// TODO revisit most of this logic once we get rid of Broker and maybe some other bits.
//  https://github.com/knative/eventing/issues/2750.
func (r *Reconciler) reconcileEventTypes(ctx context.Context, src *duckv1.Resource) error {
	current, err := r.getEventTypes(ctx, src)
	if err != nil {
		logging.FromContext(ctx).Errorw("Unable to get existing event types", zap.Error(err))
		return err
	}

	expected, err := r.makeEventTypes(ctx, src)
	if err != nil {
		return err
	}

	toCreate, toDelete := r.computeDiff(current, expected)

	for _, eventType := range toDelete {
		if err = r.eventingClientSet.EventingV1beta1().EventTypes(src.Namespace).Delete(eventType.Name, &metav1.DeleteOptions{}); err != nil {
			logging.FromContext(ctx).Errorw("Error deleting eventType", zap.Any("eventType", eventType))
			return err
		}
	}

	for _, eventType := range toCreate {
		if _, err = r.eventingClientSet.EventingV1beta1().EventTypes(src.Namespace).Create(&eventType); err != nil {
			logging.FromContext(ctx).Errorw("Error creating eventType", zap.Any("eventType", eventType))
			return err
		}
	}

	return nil
}

func (r *Reconciler) getEventTypes(ctx context.Context, src *duckv1.Resource) ([]v1beta1.EventType, error) {
	etl, err := r.eventTypeLister.EventTypes(src.Namespace).List(labels.SelectorFromSet(resources.Labels(src.Name)))
	if err != nil {
		logging.FromContext(ctx).Errorw("Unable to list event types: %v", zap.Error(err))
		return nil, err
	}
	eventTypes := make([]v1beta1.EventType, 0)
	for _, et := range etl {
		if metav1.IsControlledBy(et, src) {
			eventTypes = append(eventTypes, *et)
		}
	}
	return eventTypes, nil
}

func (r *Reconciler) makeEventTypes(ctx context.Context, src *duckv1.Resource) ([]v1beta1.EventType, error) {
	var ets []eventTypeEntry
	if v, ok := src.Annotations[eventing.EventTypesAnnotationKey]; ok {
		if err := json.Unmarshal([]byte(v), &ets); err != nil {
			// Same here, only log, can create the EventType(s) without this info.
			logging.FromContext(ctx).Errorw("Error unmarshalling EventTypes", zap.String("annotation", eventing.EventTypesAnnotationKey), zap.Error(err))
		}
	}
	eventTypes := make([]v1beta1.EventType, 0, len(ets))
	for _, et := range ets {
		schemaURL, err := apis.ParseURL(et.Schema)
		if err != nil {
			logging.FromContext(ctx).Warnw("Failed to parse schema as a URL", zap.String("schema", et.Schema), zap.Error(err))
		}
		eventType := resources.MakeEventType(&resources.EventTypeArgs{
			Resource:    src,
			CeType:      et.Type,
			CeSchema:    schemaURL,
			Description: et.Description,
		})
		eventTypes = append(eventTypes, *eventType)
	}
	return eventTypes, nil
}

func (r *Reconciler) computeDiff(current []v1beta1.EventType, expected []v1beta1.EventType) ([]v1beta1.EventType, []v1beta1.EventType) {
	toCreate := make([]v1beta1.EventType, 0)
	toDelete := make([]v1beta1.EventType, 0)
	currentMap := asMap(current, keyFromEventType)
	expectedMap := asMap(expected, keyFromEventType)

	// Iterate over the slices instead of the maps for predictable UT expectations.
	for _, e := range expected {
		if c, ok := currentMap[keyFromEventType(&e)]; !ok {
			toCreate = append(toCreate, e)
		} else {
			if !equality.Semantic.DeepEqual(e.Spec, c.Spec) {
				toDelete = append(toDelete, c)
				toCreate = append(toCreate, e)
			}
		}
	}
	// Need to check whether the current EventTypes are not in the expected map. If so, we have to delete them.
	// This could happen if the Source CO changes its broker.
	// TODO remove once we remove Broker https://github.com/knative/eventing/issues/2750
	for _, c := range current {
		if _, ok := expectedMap[keyFromEventType(&c)]; !ok {
			toDelete = append(toDelete, c)
		}
	}
	return toCreate, toDelete
}

func asMap(eventTypes []v1beta1.EventType, keyFunc func(*v1beta1.EventType) string) map[string]v1beta1.EventType {
	eventTypesAsMap := make(map[string]v1beta1.EventType)
	for _, eventType := range eventTypes {
		key := keyFunc(&eventType)
		eventTypesAsMap[key] = eventType
	}
	return eventTypesAsMap
}

// TODO we should probably use the hash of this instead. Will be revisited together with https://github.com/knative/eventing/issues/2750.
func keyFromEventType(eventType *v1beta1.EventType) string {
	return fmt.Sprintf("%s_%s_%s_%s", eventType.Spec.Type, eventType.Spec.Source, eventType.Spec.Schema, eventType.Spec.Broker)
}
