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
	eventinglisters "knative.dev/eventing/pkg/client/listers/eventing/v1"

	brokerinformer "knative.dev/eventing/pkg/client/injection/informers/eventing/v1/broker"

	"knative.dev/eventing/pkg/health"
	"knative.dev/eventing/pkg/kncloudevents"
)

type Handler struct {
	// Receiver receives incoming HTTP requests
	receiver *kncloudevents.HttpMessageReceiver
	// brokerLister gets broker objects
	brokerLister eventinglisters.BrokerLister

	logger *zap.Logger

	converter *Converter
}

func NewHandler(ctx context.Context,
	receiver *kncloudevents.HttpMessageReceiver,
	logger *zap.Logger) *Handler {

	brokerLister := brokerinformer.Get(ctx).Lister()
	return &Handler{
		receiver:     receiver,
		brokerLister: brokerLister,
		logger:       logger,
		converter:    NewConverter(ctx, logger),
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

// TODO add UID label to brokers.
func (h *Handler) getBrokerById(namespace, id string) (*eventingv1.Broker, error) {
	brokers, err := h.brokerLister.Brokers(namespace).List(labels.Everything())
	if err != nil {
		h.logger.Warn("Brokers list failed")
		return nil, err
	}
	for _, broker := range brokers {
		if string(broker.UID) == id {
			return broker, nil
		}
	}
	h.logger.Warn("Broker not found", zap.String("id", id))
	return nil, errors.NewNotFound(schema.GroupResource{}, "")
}

func (h *Handler) getBrokers(namespace string) ([]*eventingv1.Broker, error) {
	brokers, err := h.brokerLister.Brokers(namespace).List(labels.Everything())
	if err != nil {
		h.logger.Warn("Brokers list failed")
		return nil, err
	}
	return brokers, nil
}

func (h *Handler) Start(ctx context.Context) error {
	return h.receiver.StartListen(ctx, health.WithLivenessCheck(h))
}

func (h *Handler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	h.logger.Info("Called", zap.String("URI", request.RequestURI))

	// validate request method
	if request.Method != http.MethodGet {
		h.logger.Warn("unexpected request method", zap.String("method", request.Method))
		writer.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// validate request URI
	if request.RequestURI == "/" {
		writer.WriteHeader(http.StatusNotFound)
		return
	}

	requestURI := strings.Split(request.RequestURI, "/")

	if len(requestURI) != 4 && len(requestURI) != 5 {
		h.logger.Info("Malformed uri", zap.String("URI", request.RequestURI))
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	// Only allow namespaced-based queries for now.
	if requestURI[1] != "namespaces" {
		h.logger.Info("Malformed uri", zap.String("URI", request.RequestURI))
		writer.WriteHeader(http.StatusBadRequest)
		return
	}
	namespace := requestURI[2]

	if strings.HasPrefix(requestURI[3], "services") {
		h.handleServices(writer, request, requestURI, namespace)
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
		h.logger.Info("Broker retrieved", zap.String("broker", broker.Name))
		service := h.converter.ToService(broker)
		svc, err := json.Marshal(service)
		if err != nil {
			h.logger.Warn("Error marshalling service", zap.Any("service", service))
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}
		writer.Write(svc)
	} else {
		req := strings.Split(requestURI[3], "?")
		if len(req) > 2 {
			h.logger.Info("Malformed uri", zap.String("URI", request.RequestURI))
			writer.WriteHeader(http.StatusBadRequest)
			return
		} else if len(req) == 2 {
			//  /namespaces/<namespace>/services?name={name}
			r := strings.Split(req[1], "=")
			if len(r) != 2 || r[0] != "name" {
				h.logger.Info("Malformed uri", zap.String("URI", request.RequestURI))
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
			h.logger.Info("Broker retrieved", zap.String("broker", broker.Name))
			service := h.converter.ToService(broker)
			svc, err := json.Marshal(service)
			if err != nil {
				h.logger.Warn("Error marshalling service", zap.Any("service", service))
				writer.WriteHeader(http.StatusInternalServerError)
				return
			}
			writer.Write(svc)
		} else {
			// /namespaces/<namespace>/services
			brokers, err := h.getBrokers(namespace)
			if err != nil {
				writer.WriteHeader(http.StatusInternalServerError)
				return
			}
			h.logger.Info("Brokers retrieved", zap.Int("brokersCount", len(brokers)))
			services := h.converter.ToServices(brokers)
			svcs, err := json.Marshal(services)
			if err != nil {
				h.logger.Warn("Error marshalling services", zap.Any("services", services))
				writer.WriteHeader(http.StatusInternalServerError)
				return
			}
			writer.Write(svcs)
		}
	}
}
