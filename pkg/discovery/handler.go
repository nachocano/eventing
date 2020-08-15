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
	"fmt"
	"net/http"
	"net/url"
	"strings"

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

func (h *Handler) getBroker(name, namespace string) (*eventingv1.Broker, error) {
	broker, err := h.BrokerLister.Brokers(namespace).Get(name)
	if err != nil {
		h.Logger.Warn("Broker getter failed")
		return nil, err
	}
	return broker, nil
}

func (h *Handler) getEventType(name, namespace string) (*eventingv1beta1.EventType, error) {
	eventType, err := h.EventTypeLister.EventTypes(namespace).Get(name)
	if err != nil {
		h.Logger.Warn("EventType getter failed")
		return nil, err
	}
	return eventType, nil
}

func guessChannelAddress(name, namespace, domain string) string {
	url := url.URL{
		Scheme: "http",
		Host:   fmt.Sprintf("%s-kne-trigger-kn-channel.%s.svc.%s", name, namespace, domain),
		Path:   "/",
	}
	return url.String()
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
	nsBrokerName := strings.Split(request.RequestURI, "/")
	if len(nsBrokerName) != 3 {
		h.Logger.Info("Malformed uri", zap.String("URI", request.RequestURI))
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	// ctx := request.Context()
	writer.WriteHeader(500)
}
