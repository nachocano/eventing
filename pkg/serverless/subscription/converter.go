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
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	eventingv1 "knative.dev/eventing/pkg/apis/eventing/v1"
	eventinglisters "knative.dev/eventing/pkg/client/listers/eventing/v1"
	"knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"
)

type Converter struct {
	triggerLister eventinglisters.TriggerLister

	logger *zap.Logger
}

func NewConverter(triggerLister eventinglisters.TriggerLister, logger *zap.Logger) *Converter {
	return &Converter{
		triggerLister: triggerLister,
		logger:        logger,
	}
}

func (c *Converter) ToTrigger(namespacedBroker types.NamespacedName, subscription *Subscription) (*eventingv1.Trigger, error) {
	url, err := apis.ParseURL(subscription.Sink)
	if err != nil {
		return nil, err
	}
	if subscription.Protocol != "HTTP" {
		return nil, err
	}
	var attributes eventingv1.TriggerFilterAttributes
	if subscription.Filter != nil {
		if subscription.Filter.Dialect != "basic" {
			return nil, err
		}
		attributes = eventingv1.TriggerFilterAttributes{}
		for _, f := range subscription.Filter.Filters {
			if f.Type != "exact" {
				// only support exact filters
				continue
			}
			attributes[f.Property] = f.Value
		}
	}

	tr := &eventingv1.Trigger{
		ObjectMeta: v1.ObjectMeta{
			Namespace: namespacedBroker.Namespace,
			Name:      subscription.Id,
		},
		Spec: eventingv1.TriggerSpec{
			Broker: namespacedBroker.Name,
			Subscriber: duckv1.Destination{
				URI: url,
			},
		},
	}
	if attributes != nil {
		tr.Spec.Filter = &eventingv1.TriggerFilter{
			Attributes: attributes,
		}
	}
	return tr, nil
}

func (c *Converter) ToSubscriptions(triggers []*eventingv1.Trigger) []*Subscription {
	subscriptions := make([]*Subscription, len(triggers))
	for _, t := range triggers {
		subscriptions = append(subscriptions, c.ToSubscription(t))
	}
	return subscriptions
}

func (c *Converter) ToSubscription(trigger *eventingv1.Trigger) *Subscription {
	s := &Subscription{
		Id:       trigger.Name,
		Protocol: "HTTP",
		Sink:     trigger.Status.SubscriberURI.String(),
	}
	if trigger.Spec.Filter != nil {
		fd := &FilterDialect{
			Dialect: "basic",
		}
		fd.Filters = make([]Filter, len(trigger.Spec.Filter.Attributes))
		for k, v := range trigger.Spec.Filter.Attributes {
			f := Filter{
				Type:     "exact", // TODO only support exact for now
				Property: k,
				Value:    v,
			}
			fd.Filters = append(fd.Filters, f)
		}
		s.Filter = fd
	}
	return s
}
