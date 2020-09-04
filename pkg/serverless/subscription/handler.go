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

package subscription

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"k8s.io/apimachinery/pkg/types"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"k8s.io/apimachinery/pkg/labels"

	"go.uber.org/zap"
	eventingv1 "knative.dev/eventing/pkg/apis/eventing/v1"
	eventinglisters "knative.dev/eventing/pkg/client/listers/eventing/v1"

	brokerinformer "knative.dev/eventing/pkg/client/injection/informers/eventing/v1/broker"
	triggerinformer "knative.dev/eventing/pkg/client/injection/informers/eventing/v1/trigger"

	clientset "knative.dev/eventing/pkg/client/clientset/versioned"
	eventingclient "knative.dev/eventing/pkg/client/injection/client"
	"knative.dev/eventing/pkg/health"
	"knative.dev/eventing/pkg/kncloudevents"
)

type Handler struct {
	// Receiver receives incoming HTTP requests
	receiver *kncloudevents.HttpMessageReceiver
	// brokerLister gets broker objects
	brokerLister eventinglisters.BrokerLister
	// triggerLister gets trigger objects
	triggerLister eventinglisters.TriggerLister

	eventingClientSet clientset.Interface

	logger *zap.Logger

	converter *Converter
}

func NewHandler(ctx context.Context,
	receiver *kncloudevents.HttpMessageReceiver,
	logger *zap.Logger) *Handler {

	brokerLister := brokerinformer.Get(ctx).Lister()
	triggerLister := triggerinformer.Get(ctx).Lister()

	eventingClientSet := eventingclient.Get(ctx)

	converter := NewConverter(triggerLister, logger)

	return &Handler{
		receiver:          receiver,
		brokerLister:      brokerLister,
		triggerLister:     triggerLister,
		eventingClientSet: eventingClientSet,
		logger:            logger,
		converter:         converter,
	}
}

func (h *Handler) getBroker(namespace, name string) (*eventingv1.Broker, error) {
	broker, err := h.brokerLister.Brokers(namespace).Get(name)
	if err != nil {
		h.logger.Warn("Broker getter failed")
		return nil, err
	}
	return broker, nil
}

func (h *Handler) createTrigger(trigger *eventingv1.Trigger) error {
	_, err := h.eventingClientSet.EventingV1().Triggers(trigger.Namespace).Create(trigger)
	return err
}

func (h *Handler) getTriggersForBroker(namespace, name string) ([]*eventingv1.Trigger, error) {
	triggers, err := h.triggerLister.Triggers(namespace).List(labels.SelectorFromSet(map[string]string{"eventing.knative.dev/broker": name}))
	if err != nil {
		h.logger.Warn("Triggers getter failed")
		return nil, err
	}
	if len(triggers) == 0 {
		h.logger.Warn("Triggers for Broker not found", zap.String("broker", name))
		return nil, errors.NewNotFound(schema.GroupResource{}, "")
	}
	return triggers, nil
}

func (h *Handler) Start(ctx context.Context) error {
	return h.receiver.StartListen(ctx, health.WithLivenessCheck(h))
}

func (h *Handler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	h.logger.Info("Called", zap.String("URI", request.RequestURI))

	// validate request URI
	if request.RequestURI == "/" {
		writer.WriteHeader(http.StatusNotFound)
		return
	}

	nsBrokerName := strings.Split(request.RequestURI, "/")
	if len(nsBrokerName) != 3 {
		h.logger.Info("Malformed uri", zap.String("URI", request.RequestURI))
		writer.WriteHeader(http.StatusBadRequest)
		return
	}
	brokerNamespace := nsBrokerName[1]
	brokerName := nsBrokerName[2]
	brokerNamespacedName := types.NamespacedName{
		Name:      brokerName,
		Namespace: brokerNamespace,
	}

	if request.Method == http.MethodPost {
		h.handleCreateSubscription(writer, request, brokerNamespacedName)
	} else if request.Method == http.MethodGet {
		h.handleGetSubscriptions(writer, request, brokerNamespacedName)
	} else if request.Method == http.MethodDelete {
		h.handleDeleteSubscription(writer, request, brokerNamespacedName)
	} else if request.Method == http.MethodPut {
		h.handleUpdateSubscription(writer, request, brokerNamespacedName)
	}

}

func (h *Handler) handleCreateSubscription(writer http.ResponseWriter, request *http.Request, namespacedBroker types.NamespacedName) {
	var s Subscription
	err := json.NewDecoder(request.Body).Decode(&s)
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	trigger, err := h.converter.ToTrigger(namespacedBroker, &s)
	if err != nil {
		h.logger.Error("Error converting Subscription to Trigger", zap.Error(err), zap.Any("subscription", s))
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	broker, err := h.getBroker(namespacedBroker.Namespace, namespacedBroker.Name)
	if err != nil {
		if errors.IsNotFound(err) {
			writer.WriteHeader(http.StatusNotFound)
			return
		}
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	h.logger.Info("Broker retrieved", zap.String("broker", broker.Name))

	err = h.createTrigger(trigger)
	if err != nil {
		if errors.IsAlreadyExists(err) {
			writer.WriteHeader(http.StatusConflict)
		} else {
			writer.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	writer.WriteHeader(http.StatusCreated)
}

func (h *Handler) handleGetSubscriptions(writer http.ResponseWriter, request *http.Request, namespacedBroker types.NamespacedName) {
	triggers, err := h.getTriggersForBroker(namespacedBroker.Namespace, namespacedBroker.Name)
	if err != nil {
		if errors.IsNotFound(err) {
			writer.WriteHeader(http.StatusNotFound)
			return
		}
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	h.logger.Info("Triggers retrieved", zap.Int("triggersCount", len(triggers)))
	subscriptions := h.converter.ToSubscriptions(triggers)
	subs, err := json.Marshal(subscriptions)
	if err != nil {
		h.logger.Warn("Error marshalling subscriptions", zap.Any("subscriptions", subscriptions))
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	writer.Write(subs)
}

func (h *Handler) handleDeleteSubscription(writer http.ResponseWriter, request *http.Request, namespacedBroker types.NamespacedName) {
	// TODO
}

func (h *Handler) handleUpdateSubscription(writer http.ResponseWriter, request *http.Request, namespacedBroker types.NamespacedName) {
	// TODO
}
