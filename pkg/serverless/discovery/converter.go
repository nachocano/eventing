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

	"knative.dev/eventing/pkg/reconciler/names"
	"knative.dev/pkg/system"

	"go.uber.org/zap"
	eventingv1 "knative.dev/eventing/pkg/apis/eventing/v1"
	eventtypeinformer "knative.dev/eventing/pkg/client/injection/informers/eventing/v1/eventtype"
	schemainformer "knative.dev/eventing/pkg/client/injection/informers/eventing/v1alpha1/schema"
	eventinglisters "knative.dev/eventing/pkg/client/listers/eventing/v1"
	eventingv1alpha1listers "knative.dev/eventing/pkg/client/listers/eventing/v1alpha1"
)

type Converter struct {
	eventTypeLister eventinglisters.EventTypeLister
	schemaLister    eventingv1alpha1listers.SchemaLister

	logger *zap.Logger
}

func NewConverter(ctx context.Context, logger *zap.Logger) *Converter {
	eventTypeLister := eventtypeinformer.Get(ctx).Lister()
	schemaLister := schemainformer.Get(ctx).Lister()
	return &Converter{
		eventTypeLister: eventTypeLister,
		schemaLister:    schemaLister,
		logger:          logger,
	}
}

func (c *Converter) ToServices(brokers []*eventingv1.Broker) []*Service {
	services := make([]*Service, len(brokers))
	for _, b := range brokers {
		services = append(services, c.ToService(b))
	}
	return services
}

func (c *Converter) ToService(broker *eventingv1.Broker) *Service {
	svc := &Service{
		Id:              string(broker.UID),
		Name:            broker.Name,
		Url:             fmt.Sprintf("http://%s/services/%s", names.ServiceHostName("discovery", system.Namespace()), string(broker.UID)),
		SpecVersions:    []string{"0.3", "1.0"},
		SubscriptionUrl: fmt.Sprintf("http://%s/%s/%s", names.ServiceHostName("subscription", system.Namespace()), broker.Namespace, broker.Name),
		Protocols:       []string{"HTTP"},
	}

	events := make([]Event, 0)
	// TODO use the names and a fieldSelector and the clientSet to do a single query to the ApiServer?
	//  Although the lister uses the cache and might be more efficient
	for _, et := range broker.Spec.EventTypes {
		eventType, err := c.eventTypeLister.EventTypes(broker.Namespace).Get(et.Name)
		if err != nil {
			c.logger.Error("Error retrieving eventType", zap.Error(err), zap.String("eventType", et.Name), zap.Any("broker", broker))
			// TODO should return an error?
			continue
		}
		var dataSchema, dataSchemaType, dataSchemaContent string
		if eventType.Spec.Schema != nil {
			dataSchema = eventType.Spec.Schema.URI.String()

			if eventType.Spec.Schema.Ref != nil {
				schema, err := c.schemaLister.Schemas(broker.Namespace).Get(eventType.Spec.Schema.Ref.Name)
				if err != nil {
					c.logger.Error("Error retrieving schema", zap.Error(err), zap.String("schema", eventType.Spec.Schema.Ref.Name), zap.Any("broker", broker))
					// TODO should return an error?
					continue
				}
				dataSchemaType = schema.Spec.Format
				for _, v := range schema.Spec.Versions {
					if v.Version == eventType.Spec.Schema.Ref.Version {
						dataSchemaContent = v.Body
						break
					}
				}
			}
		}
		et := Event{
			Type:              eventType.Spec.Type,
			Description:       eventType.Spec.Description,
			DataContentType:   eventType.Spec.ContentType,
			DataSchema:        dataSchema,
			DataSchemaType:    dataSchemaType,
			DataSchemaContent: dataSchemaContent,
			SourceTemplate:    eventType.Spec.SourceTemplate,
			Extensions:        eventType.Spec.Extensions,
		}
		events = append(events, et)
	}
	svc.Events = events
	return svc
}
