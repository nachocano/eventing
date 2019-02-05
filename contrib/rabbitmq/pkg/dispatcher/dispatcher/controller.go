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

package dispatcher

import (
	"context"
	"sync"

	"sigs.k8s.io/controller-runtime/pkg/event"

	eventingv1alpha1 "github.com/knative/eventing/pkg/apis/eventing/v1alpha1"
	"github.com/knative/eventing/pkg/provisioners"
	"go.uber.org/zap"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

const (
	// controllerAgentName is the string used by this controller to identify itself when creating
	// events.
	controllerAgentName = "rabbitmq-channel-dispatcher"
)

// New returns a Controller that represents the dispatcher portion (messages from RabbitMq are
// sent into the cluster) of the RabbitMq dispatcher. We use a reconcile loop to watch all
// Channels and notice changes to them.
func New(mgr manager.Manager, logger *zap.Logger, stopCh <-chan struct{}) (controller.Controller, error) {
	// reconcileChan is used when the dispatcher itself needs to force reconciliation of a Channel.
	reconcileChan := make(chan event.GenericEvent)

	// Setup a new controller to pull messages from RabbitMq for Channels that belong to this
	// Cluster Provisioner (rabbitmq).
	r := &reconciler{
		recorder: mgr.GetRecorder(controllerAgentName),
		logger:   logger,

		dispatcher:    provisioners.NewMessageDispatcher(logger.Sugar()),
		reconcileChan: reconcileChan,

		subscriptionsLock: sync.Mutex{},
		subscriptions:     map[channelName]map[subscriptionName]context.CancelFunc{},
	}

	c, err := controller.New(controllerAgentName, mgr, controller.Options{
		Reconciler: r,
	})
	if err != nil {
		logger.Error("Unable to create controller.", zap.Error(err))
		return nil, err
	}

	// Watch Channels.
	err = c.Watch(&source.Kind{
		Type: &eventingv1alpha1.Channel{},
	}, &handler.EnqueueRequestForObject{})
	if err != nil {
		logger.Error("Unable to watch Channels.", zap.Error(err), zap.Any("type", &eventingv1alpha1.Channel{}))
		return nil, err
	}

	// The PubSub library may fail when receiving messages. If it does so, then we need to reconcile
	// that Channel again.
	// TODO Once https://github.com/kubernetes-sigs/controller-runtime/issues/103 is fixed, switch
	// to:
	// err = c.Watch(&source.Channel{
	// 	Source: reconcileChan,
	// }, &handler.EnqueueRequestForOwner{})
	src := &source.Channel{
		Source: reconcileChan,
	}
	err = src.InjectStopChannel(stopCh)
	if err != nil {
		logger.Error("Unable to inject the stop channel", zap.Error(err))
		return nil, err
	}
	err = c.Watch(src, &handler.EnqueueRequestForObject{})
	if err != nil {
		logger.Error("Unable to watch the reconcile Channel", zap.Error(err))
		return nil, err
	}

	return c, nil
}
