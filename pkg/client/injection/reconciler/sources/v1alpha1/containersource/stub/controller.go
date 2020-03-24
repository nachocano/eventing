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

// Code generated by injection-gen. DO NOT EDIT.

package containersource

import (
	context "context"

	containersource "knative.dev/eventing/pkg/client/injection/informers/sources/v1alpha1/containersource"
	v1alpha1containersource "knative.dev/eventing/pkg/client/injection/reconciler/sources/v1alpha1/containersource"
	configmap "knative.dev/pkg/configmap"
	controller "knative.dev/pkg/controller"
	logging "knative.dev/pkg/logging"
)

// TODO: PLEASE COPY AND MODIFY THIS FILE AS A STARTING POINT

// NewController creates a Reconciler for ContainerSource and returns the result of NewImpl.
func NewController(
	ctx context.Context,
	cmw configmap.Watcher,
) *controller.Impl {
	logger := logging.FromContext(ctx)

	containersourceInformer := containersource.Get(ctx)

	// TODO: setup additional informers here.

	r := &Reconciler{}
	impl := v1alpha1containersource.NewImpl(ctx, r)

	logger.Info("Setting up event handlers.")

	containersourceInformer.Informer().AddEventHandler(controller.HandleAll(impl.Enqueue))

	// TODO: add additional informer event handlers here.

	return impl
}
