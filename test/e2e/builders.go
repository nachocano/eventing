/*
Copyright 2019 The Knative Authors
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

package e2e

import (
	eventingv1alpha1 "github.com/knative/eventing/pkg/apis/eventing/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Broker builder.
type BrokerBuilder struct {
	*eventingv1alpha1.Broker
}

func NewBrokerBuilder(name, namespace string) *BrokerBuilder {
	broker := &eventingv1alpha1.Broker{
		TypeMeta: metav1.TypeMeta{
			APIVersion: eventingv1alpha1.SchemeGroupVersion.String(),
			Kind:       "Broker",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: eventingv1alpha1.BrokerSpec{},
	}

	return &BrokerBuilder{
		Broker: broker,
	}
}

func (b *BrokerBuilder) Build() *eventingv1alpha1.Broker {
	return b.Broker.DeepCopy()
}

// Trigger builder.
type TriggerBuilder struct {
	*eventingv1alpha1.Trigger
}

func NewTriggerBuilder(name, namespace string) *TriggerBuilder {
	trigger := &eventingv1alpha1.Trigger{
		TypeMeta: metav1.TypeMeta{
			APIVersion: eventingv1alpha1.SchemeGroupVersion.String(),
			Kind:       "Trigger",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: eventingv1alpha1.TriggerSpec{
			// Bind to 'default' Broker by default
			Broker: "default",
			Filter: &eventingv1alpha1.TriggerFilter{
				// Create a Any filter by default.
				SourceAndType: &eventingv1alpha1.TriggerFilterSourceAndType{
					Source: eventingv1alpha1.TriggerAnyFilter,
					Type:   eventingv1alpha1.TriggerAnyFilter,
				},
			},
		},
	}

	return &TriggerBuilder{
		Trigger: trigger,
	}
}

func (t *TriggerBuilder) Build() *eventingv1alpha1.Trigger {
	return t.Trigger.DeepCopy()
}

func (t *TriggerBuilder) Type(eventType string) *TriggerBuilder {
	t.Trigger.Spec.Filter.SourceAndType.Type = eventType
	return t
}

func (t *TriggerBuilder) Source(eventSource string) *TriggerBuilder {
	t.Trigger.Spec.Filter.SourceAndType.Source = eventSource
	return t
}

func (t *TriggerBuilder) Broker(brokerName string) *TriggerBuilder {
	t.Trigger.Spec.Broker = brokerName
	return t
}

func (t *TriggerBuilder) Subscriber(ref *corev1.ObjectReference) *TriggerBuilder {
	t.Trigger.Spec.Subscriber.Ref = ref
	return t
}

func (t *TriggerBuilder) SubscriberSvc(svcName string) *TriggerBuilder {
	t.Trigger.Spec.Subscriber.Ref = &corev1.ObjectReference{
		APIVersion: corev1.SchemeGroupVersion.String(),
		Kind:       "Service",
		Name:       svcName,
		Namespace:  t.Trigger.GetNamespace(),
	}
	return t
}
