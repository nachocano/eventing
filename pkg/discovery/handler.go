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

package discovery

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"k8s.io/apimachinery/pkg/labels"

	"go.uber.org/zap"
	eventingv1 "knative.dev/eventing/pkg/apis/eventing/v1"
	eventingv1beta1 "knative.dev/eventing/pkg/apis/eventing/v1beta1"
	eventinglisters "knative.dev/eventing/pkg/client/listers/eventing/v1"
	eventingv1beta1listers "knative.dev/eventing/pkg/client/listers/eventing/v1beta1"
	"knative.dev/eventing/pkg/health"
	"knative.dev/eventing/pkg/kncloudevents"
)

const (
	// noDuration signals that the dispatch step hasn't started
	noDuration = -1
)

type Handler struct {
	// Receiver receives incoming HTTP requests
	Receiver *kncloudevents.HttpMessageReceiver
	// BrokerLister gets broker objects
	BrokerLister eventinglisters.BrokerLister
	// EventTypeLister gets event types objects
	EventTypeLister eventingv1beta1listers.EventTypeLister

	Logger *zap.Logger
}

func (h *Handler) getBroker(namespace, name string) (*eventingv1.Broker, error) {
	broker, err := h.BrokerLister.Brokers(namespace).Get(name)
	if err != nil {
		h.Logger.Warn("Broker getter failed")
		return nil, err
	}
	return broker, nil
}

// TODO add UID label to brokers.
func (h *Handler) getBrokerById(namespace, id string) (*eventingv1.Broker, error) {
	brokers, err := h.BrokerLister.Brokers(namespace).List(labels.Everything())
	if err != nil {
		h.Logger.Warn("Brokers list failed")
		return nil, err
	}
	for _, broker := range brokers {
		if string(broker.UID) == id {
			return broker, nil
		}
	}
	h.Logger.Warn("Broker not found", zap.String("id", id))
	return nil, errors.NewNotFound(schema.GroupResource{}, "")
}

func (h *Handler) getBrokers(namespace string) ([]*eventingv1.Broker, error) {
	brokers, err := h.BrokerLister.Brokers(namespace).List(labels.Everything())
	if err != nil {
		h.Logger.Warn("Brokers list failed")
		return nil, err
	}
	return brokers, nil
}

// TODO this could actually return a list. Correct spec.
// TODO add type label to EventTypes
func (h *Handler) getEventType(namespace, type_ string) (*eventingv1beta1.EventType, error) {
	eventTypes, err := h.EventTypeLister.EventTypes(namespace).List(labels.Everything())
	if err != nil {
		h.Logger.Warn("EventTypes list failed")
		return nil, err
	}
	for _, eventType := range eventTypes {
		if eventType.Spec.Type == type_ {
			return eventType, nil
		}
	}
	h.Logger.Warn("EventType not found", zap.String("type", type_))
	return nil, errors.NewNotFound(schema.GroupResource{}, "")
}

func (h *Handler) getEventTypes(namespace string) ([]*eventingv1beta1.EventType, error) {
	eventTypes, err := h.EventTypeLister.EventTypes(namespace).List(labels.Everything())
	if err != nil {
		h.Logger.Warn("EventTypes list failed")
		return nil, err
	}
	return eventTypes, nil
}

func (h *Handler) getEventTypesMatching(namespace, typeSubstring string) ([]*eventingv1beta1.EventType, error) {
	eventTypes, err := h.EventTypeLister.EventTypes(namespace).List(labels.Everything())
	if err != nil {
		h.Logger.Warn("EventTypes list failed")
		return nil, err
	}
	ets := make([]*eventingv1beta1.EventType, 0)
	for _, eventType := range eventTypes {
		if strings.Contains(strings.ToLower(eventType.Spec.Type), strings.ToLower(typeSubstring)) {
			ets = append(ets, eventType)
		}
	}
	if len(ets) == 0 {
		h.Logger.Warn("EventTypes not found", zap.String("typeSubstring", typeSubstring))
		return nil, errors.NewNotFound(schema.GroupResource{}, "")
	}
	return ets, nil
}

func (h *Handler) Start(ctx context.Context) error {
	return h.Receiver.StartListen(ctx, health.WithLivenessCheck(h))
}

func (h *Handler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	h.Logger.Info("Called", zap.String("URI", request.RequestURI))

	// validate request method
	if request.Method != http.MethodGet {
		h.Logger.Warn("unexpected request method", zap.String("method", request.Method))
		writer.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// validate request URI
	if request.RequestURI == "/" {
		writer.WriteHeader(http.StatusNotFound)
		return
	}

	requestURI := strings.Split(request.RequestURI, "/")

	for i, req := range requestURI {
		h.Logger.Info("URI", zap.Int("idx", i), zap.String("req", req))
	}

	if len(requestURI) != 4 && len(requestURI) != 5 {
		h.Logger.Info("Malformed uri", zap.String("URI", request.RequestURI))
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	// Only allow namespaced-based queries for now.
	if requestURI[1] != "namespaces" {
		h.Logger.Info("Malformed uri", zap.String("URI", request.RequestURI))
		writer.WriteHeader(http.StatusBadRequest)
		return
	}
	namespace := requestURI[2]

	if strings.HasPrefix(requestURI[3], "services") {
		h.handleServices(writer, request, requestURI, namespace)
	} else if strings.HasPrefix(requestURI[3], "types") {
		h.handleTypes(writer, request, requestURI, namespace)
	} else {
		writer.WriteHeader(http.StatusNotFound)
	}
}

func (h *Handler) handleServices(writer http.ResponseWriter, request *http.Request, requestURI []string, namespace string) {
	if len(requestURI) == 5 {
		// /namespaces/<namespace>/services/{id}
		broker, err := h.getBrokerById(namespace, requestURI[4])
		if err != nil {
			if errors.IsNotFound(err) {
				writer.WriteHeader(http.StatusNotFound)
				return
			}
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}
		h.Logger.Info("Broker retrieved", zap.String("broker", broker.Name))
		b, err := json.Marshal(broker)
		if err != nil {
			h.Logger.Warn("Error marshalling broker", zap.Any("broker", broker))
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}
		writer.Write(b)
	} else {
		req := strings.Split(requestURI[3], "?")
		if len(req) > 2 {
			h.Logger.Info("Malformed uri", zap.String("URI", request.RequestURI))
			writer.WriteHeader(http.StatusBadRequest)
			return
		} else if len(req) == 2 {
			//  /namespaces/<namespace>/services?name={name}
			r := strings.Split(req[1], "=")
			if len(r) != 2 || r[0] != "name" {
				h.Logger.Info("Malformed uri", zap.String("URI", request.RequestURI))
				writer.WriteHeader(http.StatusBadRequest)
				return
			}
			broker, err := h.getBroker(namespace, r[1])
			if err != nil {
				if errors.IsNotFound(err) {
					writer.WriteHeader(http.StatusNotFound)
					return
				}
				writer.WriteHeader(http.StatusInternalServerError)
				return
			}
			h.Logger.Info("Broker retrieved", zap.String("broker", broker.Name))
			b, err := json.Marshal(broker)
			if err != nil {
				h.Logger.Warn("Error marshalling broker", zap.Any("broker", broker))
				writer.WriteHeader(http.StatusInternalServerError)
				return
			}
			writer.Write(b)
		} else {
			// /namespaces/<namespace>/services
			brokers, err := h.getBrokers(namespace)
			if err != nil {
				writer.WriteHeader(http.StatusInternalServerError)
				return
			}
			h.Logger.Info("Brokers retrieved", zap.Int("brokersCount", len(brokers)))
			bs, err := json.Marshal(brokers)
			if err != nil {
				h.Logger.Warn("Error marshalling brokers", zap.Any("brokers", bs))
				writer.WriteHeader(http.StatusInternalServerError)
				return
			}
			writer.Write(bs)
		}
	}
}

func (h *Handler) handleTypes(writer http.ResponseWriter, request *http.Request, requestURI []string, namespace string) {
	if len(requestURI) == 5 {
		// /namespaces/<namespace>/types/{type}
		et, err := h.getEventType(namespace, requestURI[4])
		if err != nil {
			if errors.IsNotFound(err) {
				writer.WriteHeader(http.StatusNotFound)
				return
			}
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}
		h.Logger.Info("EventType retrieved", zap.String("eventType", et.Name))
		e, err := json.Marshal(et)
		if err != nil {
			h.Logger.Warn("Error marshalling EventType", zap.Any("eventType", et))
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}
		writer.Write(e)
	} else {
		req := strings.Split(requestURI[3], "?")
		if len(req) > 2 {
			h.Logger.Info("Malformed uri", zap.String("URI", request.RequestURI))
			writer.WriteHeader(http.StatusBadRequest)
			return
		} else if len(req) == 2 {
			// /namespaces/<namespace>/types?matching={name}
			r := strings.Split(req[1], "=")
			if len(r) != 2 || r[0] != "matching" {
				h.Logger.Info("Malformed uri", zap.String("URI", request.RequestURI))
				writer.WriteHeader(http.StatusBadRequest)
				return
			}
			ets, err := h.getEventTypesMatching(namespace, r[1])
			if err != nil {
				if errors.IsNotFound(err) {
					writer.WriteHeader(http.StatusNotFound)
					return
				}
				writer.WriteHeader(http.StatusInternalServerError)
				return
			}
			h.Logger.Info("EventTypes retrieved", zap.Int("eventTypesCount", len(ets)))
			es, err := json.Marshal(ets)
			if err != nil {
				h.Logger.Warn("Error marshalling EventTypes", zap.Any("eventTypes", es))
				writer.WriteHeader(http.StatusInternalServerError)
				return
			}
			writer.Write(es)
		} else {
			// /namespaces/<namespace>/types
			ets, err := h.getEventTypes(namespace)
			if err != nil {
				writer.WriteHeader(http.StatusInternalServerError)
				return
			}
			h.Logger.Info("EventTypes retrieved", zap.Int("eventTypesCount", len(ets)))
			es, err := json.Marshal(ets)
			if err != nil {
				h.Logger.Warn("Error marshalling EventTypes", zap.Any("eventType", ets))
				writer.WriteHeader(http.StatusInternalServerError)
				return
			}
			writer.Write(es)
		}
	}
}
