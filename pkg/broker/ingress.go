/*
 * Copyright 2018 The Knative Authors
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

package broker

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"go.uber.org/zap"

	"k8s.io/apimachinery/pkg/util/validation"

	"github.com/cloudevents/sdk-go/pkg/cloudevents"

	eventingv1alpha1 "github.com/knative/eventing/pkg/apis/eventing/v1alpha1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	allowAny             = "allow_any"
	allowRegisteredTypes = "allow_registered"
	autoCreate           = "auto_create"
)

var (
	// Only allow alphanumeric, '-' or '.'.
	validChars = regexp.MustCompile(`[^-\.a-z0-9]+`)
	// EventType not found error.
	notFound = k8serrors.NewNotFound(eventingv1alpha1.Resource("eventtype"), "")
)

type IngressPolicy interface {
	AllowEvent(event *cloudevents.Event) bool
}

type Any struct{}

type Registered struct {
	logger    *zap.SugaredLogger
	client    client.Client
	namespace string
	broker    string
}

type AutoCreate struct {
	logger    *zap.SugaredLogger
	client    client.Client
	namespace string
	broker    string
}

func NewIngressPolicy(logger *zap.Logger, client client.Client, namespace, broker, policy string) IngressPolicy {
	return newIngressPolicy(logger.Sugar(), client, namespace, broker, policy)
}

func newIngressPolicy(logger *zap.SugaredLogger, client client.Client, namespace, broker, policy string) IngressPolicy {
	switch policy {
	case allowRegisteredTypes:
		return &Registered{
			logger:    logger,
			client:    client,
			namespace: namespace,
			broker:    broker,
		}
	case autoCreate:
		return &AutoCreate{
			logger:    logger,
			client:    client,
			namespace: namespace,
			broker:    broker,
		}
	case allowAny:
		return &Any{}
	default:
		return &Any{}
	}
}

func (p *Any) AllowEvent(event *cloudevents.Event) bool {
	return true
}

func (p *Registered) AllowEvent(event *cloudevents.Event) bool {
	_, err := getEventType(p.client, event, p.namespace, p.broker)
	if k8serrors.IsNotFound(err) {
		p.logger.Warnf("EventType not found, rejecting spec.type %q, %v", event.Type(), err)
		return false
	} else if err != nil {
		p.logger.Errorf("Error retrieving EventType, rejecting spec.type %q, %v", event.Type(), err)
		return false
	}
	return true
}

func (p *AutoCreate) AllowEvent(event *cloudevents.Event) bool {
	_, err := getEventType(p.client, event, p.namespace, p.broker)
	if k8serrors.IsNotFound(err) {
		p.logger.Infof("EventType not found, creating spec.type %q", event.Type())
		eventType := makeEventType(event, p.namespace, p.broker)
		err := p.client.Create(context.TODO(), eventType)
		if err != nil {
			p.logger.Errorf("Error creating EventType, spec.type %q, %v", event.Type(), err)
		}
	} else if err != nil {
		p.logger.Errorf("Error retrieving EventType, spec.type %q, %v", event.Type(), err)
	}
	return true
}

func getEventType(c client.Client, event *cloudevents.Event, namespace, broker string) (*eventingv1alpha1.EventType, error) {
	opts := &client.ListOptions{
		Namespace: namespace,
		// Set Raw because if we need to get more than one page, then we will put the continue token
		// into opts.Raw.Continue.
		Raw: &metav1.ListOptions{},
	}

	ctx := context.TODO()

	for {
		etl := &eventingv1alpha1.EventTypeList{}
		err := c.List(ctx, opts, etl)
		if err != nil {
			return nil, err
		}
		for _, et := range etl.Items {
			if et.Spec.Broker == broker {
				// TODO what about source.
				if et.Spec.Type == event.Type() && et.Spec.Schema == event.SchemaURL() {
					return &et, nil
				}
			}
		}
		if etl.Continue != "" {
			opts.Raw.Continue = etl.Continue
		} else {
			return nil, notFound
		}
	}
}

// makeEventType generates, but does not create an EventType from the given cloudevents.Event.
func makeEventType(event *cloudevents.Event, namespace, broker string) *eventingv1alpha1.EventType {
	cloudEventType := event.Type()
	return &eventingv1alpha1.EventType{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: fmt.Sprintf("%s-", toDNS1123Subdomain(cloudEventType)),
			Namespace:    namespace,
		},
		Spec: eventingv1alpha1.EventTypeSpec{
			Type:   cloudEventType,
			Source: event.Source(),
			Schema: event.SchemaURL(),
			Broker: broker,
		},
	}
}

func toDNS1123Subdomain(cloudEventType string) string {
	// If it is not a valid DNS1123 subdomain, make it a valid one.
	if msgs := validation.IsDNS1123Subdomain(cloudEventType); len(msgs) != 0 {
		// If the length exceeds the max, cut it and leave some room for the generated UUID.
		if len(cloudEventType) > validation.DNS1123SubdomainMaxLength {
			cloudEventType = cloudEventType[:validation.DNS1123SubdomainMaxLength-10]
		}
		cloudEventType = strings.ToLower(cloudEventType)
		cloudEventType = validChars.ReplaceAllString(cloudEventType, "")
		// Only start/end with alphanumeric.
		cloudEventType = strings.Trim(cloudEventType, "-.")
	}
	return cloudEventType
}
