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

	"k8s.io/apimachinery/pkg/runtime/schema"
	"knative.dev/pkg/injection"

	"go.uber.org/zap"
	"k8s.io/client-go/tools/cache"

	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/logging"

	pkgreconciler "knative.dev/pkg/reconciler"

	eventingclient "knative.dev/eventing/pkg/client/injection/client"
	resourceinformer "knative.dev/eventing/pkg/client/injection/ducks/duck/v1/resource"
	eventtypeinformer "knative.dev/eventing/pkg/client/injection/informers/eventing/v1beta1/eventtype"
)

const (
	// ReconcilerName is the name of the reconciler.
	ReconcilerName = "Producer"
)

// NewController returns a function that initializes the controller and
// Registers event handlers to enqueue events
func NewController(gvr schema.GroupVersionResource, gvk schema.GroupVersionKind) injection.ControllerConstructor {
	return func(ctx context.Context,
		cmw configmap.Watcher,
	) *controller.Impl {
		logger := logging.FromContext(ctx)
		eventTypeInformer := eventtypeinformer.Get(ctx)
		resourceDuckInformer := resourceinformer.Get(ctx)

		resourceInformer, resourceLister, err := resourceDuckInformer.Get(gvr)
		if err != nil {
			logger.Errorw("Error getting resource informer", zap.String("GVR", gvr.String()), zap.Error(err))
			return nil
		}

		r := &Reconciler{
			eventingClientSet: eventingclient.Get(ctx),
			eventTypeLister:   eventTypeInformer.Lister(),
			resourceLister:    resourceLister,
			gvr:               gvr,
		}
		impl := controller.NewImpl(r, logger, ReconcilerName)

		logger.Info("Setting up event handlers")
		resourceInformer.AddEventHandler(cache.FilteringResourceEventHandler{
			FilterFunc: pkgreconciler.LabelFilterFunc("producers.knative.dev/producer", "true", false),
			Handler:    controller.HandleAll(impl.Enqueue),
		})

		eventTypeInformer.Informer().AddEventHandler(cache.FilteringResourceEventHandler{
			FilterFunc: controller.FilterControllerGVK(gvk),
			Handler:    controller.HandleAll(impl.EnqueueControllerOf),
		})

		return impl
	}
}
