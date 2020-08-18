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
	"go.uber.org/zap"
	eventingv1 "knative.dev/eventing/pkg/apis/eventing/v1"
	eventinglisters "knative.dev/eventing/pkg/client/listers/eventing/v1"
)

type Converter struct {
	eventTypeLister eventinglisters.EventTypeLister

	logger *zap.Logger
}

func NewConverter(eventTypeLister eventinglisters.EventTypeLister, logger *zap.Logger) *Converter {
	return &Converter{
		eventTypeLister: eventTypeLister,
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
		Url:             "",
		SpecVersions:    []string{"0.3", "1.0"},
		SubscriptionUrl: "",
		Protocols:       []string{"HTTP"},
	}

	types := make([]Type, 0)
	// TODO use the names and a fieldSelector and the clientSet to do a single query to the ApiServer?
	//  Although the lister uses the cache and might be more efficient
	for _, et := range broker.Spec.EventTypes {
		eventType, err := c.eventTypeLister.EventTypes(broker.Namespace).Get(et.Name)
		if err != nil {
			c.logger.Error("Error retrieving eventType", zap.String("eventType", et.Name), zap.Any("broker", broker))
			// TODO should return an error?
			continue
		}
		et := Type{
			Type:              eventType.Spec.Type,
			Description:       eventType.Spec.Description,
			DataContentType:   eventType.Spec.ContentType,
			DataSchema:        eventType.Spec.Schema.String(),
			DataSchemaType:    eventType.Spec.SchemaDataType,
			DataSchemaContent: eventType.Spec.SchemaData,
			SourceTemplate:    eventType.Spec.SourceTemplate,
			Extensions:        eventType.Spec.Extensions,
		}
		types = append(types, et)
	}
	svc.Types = types
	return svc
}
