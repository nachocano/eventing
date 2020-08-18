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

func (c *Converter) ToTriggers(subscriptions []Subscription) []eventingv1.Trigger {
	triggers := make([]eventingv1.Trigger, len(subscriptions))
	for _, s := range subscriptions {
		triggers = append(triggers, c.ToTrigger(&s))
	}
	return triggers
}

func (c *Converter) ToTrigger(subscription *Subscription) eventingv1.Trigger {
	tr := eventingv1.Trigger{
		// TODO populate
	}
	return tr
}
