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

package testing

import (
	"context"
	"time"

	v1 "k8s.io/api/apps/v1"

	"github.com/knative/eventing/pkg/apis/eventing/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// BrokerOption enables further configuration of a Broker.
type BrokerOption func(*v1alpha1.Broker)

// NewBroker creates a Broker with BrokerOptions.
func NewBroker(name, namespace string, o ...BrokerOption) *v1alpha1.Broker {
	b := &v1alpha1.Broker{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "eventing.knative.dev/v1alpha1",
			Kind:       "Broker",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
		Spec: v1alpha1.BrokerSpec{
			ChannelTemplate: &v1alpha1.ChannelSpec{},
		},
	}
	for _, opt := range o {
		opt(b)
	}
	b.SetDefaults(context.Background())
	return b
}

// WithInitBrokerConditions initializes the Broker's conditions.
func WithInitBrokerConditions(b *v1alpha1.Broker) {
	b.Status.InitializeConditions()
}

func WithBrokerDeletionTimestamp(b *v1alpha1.Broker) {
	t := metav1.NewTime(time.Unix(1e9, 0))
	b.ObjectMeta.SetDeletionTimestamp(&t)
}

// WithBrokerChannelProvisioner sets the Broker's ChannelTemplate provisioner.
func WithBrokerChannelProvisioner(provisioner *corev1.ObjectReference) BrokerOption {
	return func(b *v1alpha1.Broker) {
		b.Spec.ChannelTemplate.Provisioner = provisioner
	}
}

// WithBrokerAddress sets the Broker's address.
func WithBrokerAddress(address string) BrokerOption {
	return func(b *v1alpha1.Broker) {
		b.Status.SetAddress(address)
	}
}

// MarkTriggerChannelFailed calls .Status.MarkTriggerChannelFailed on the Broker.
func MarkTriggerChannelFailed(reason, format, arg string) BrokerOption {
	return func(b *v1alpha1.Broker) {
		b.Status.MarkTriggerChannelFailed(reason, format, arg)
	}
}

// MarkFilterFailed calls .Status.MarkFilterFailed on the Broker.
func MarkFilterFailed(reason, format, arg string) BrokerOption {
	return func(b *v1alpha1.Broker) {
		b.Status.MarkFilterFailed(reason, format, arg)
	}
}

// MarkIngressFailed calls .Status.MarkIngressFailed on the Broker.
func MarkIngressFailed(reason, format, arg string) BrokerOption {
	return func(b *v1alpha1.Broker) {
		b.Status.MarkIngressFailed(reason, format, arg)
	}
}

// MarkIngressChannelFailed calls .Status.MarkIngressChannelFailed on the Broker.
func MarkIngressChannelFailed(reason, format, arg string) BrokerOption {
	return func(b *v1alpha1.Broker) {
		b.Status.MarkIngressChannelFailed(reason, format, arg)
	}
}

// PropagateTriggerChannelReadiness calls .Status.PropagateTriggerChannelReadiness on the Broker.
func PropagateTriggerChannelReadiness(cs *v1alpha1.ChannelStatus) BrokerOption {
	return func(b *v1alpha1.Broker) {
		b.Status.PropagateTriggerChannelReadiness(cs)
	}
}

// PropagateFilterDeploymentAvailability calls .Status.PropagateFilterDeploymentAvailability on the Broker.
func PropagateFilterDeploymentAvailability(d *v1.Deployment) BrokerOption {
	return func(b *v1alpha1.Broker) {
		b.Status.PropagateFilterDeploymentAvailability(d)
	}
}

// PropagateIngressDeploymentAvailability calls .Status.PropagateIngressDeploymentAvailability on the Broker.
func PropagateIngressDeploymentAvailability(d *v1.Deployment) BrokerOption {
	return func(b *v1alpha1.Broker) {
		b.Status.PropagateIngressDeploymentAvailability(d)
	}
}

// PropagateIngressDeploymentAvailability calls .Status.PropagateIngressChannelReadiness on the Broker.
func PropagateIngressChannelReadiness(cs *v1alpha1.ChannelStatus) BrokerOption {
	return func(b *v1alpha1.Broker) {
		b.Status.PropagateIngressChannelReadiness(cs)
	}
}

// PropagateIngressSubscriptionReadiness calls .Status.PropagateIngressSubscriptionReadiness on the Broker.
func PropagateIngressSubscriptionReadiness(ss *v1alpha1.SubscriptionStatus) BrokerOption {
	return func(b *v1alpha1.Broker) {
		b.Status.PropagateIngressSubscriptionReadiness(ss)
	}
}
