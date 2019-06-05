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

	"github.com/knative/eventing/contrib/natss/pkg/apis/messaging/v1alpha1"
	"github.com/knative/pkg/apis"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	//	"k8s.io/apimachinery/pkg/types"
)

// NatssChannelOption enables further configuration of a NatssChannel.
type NatssChannelOption func(*v1alpha1.NatssChannel)

// NewNatssChannel creates an NatssChannel with NatssChannelOptions.
func NewNatssChannel(name, namespace string, imcopt ...NatssChannelOption) *v1alpha1.NatssChannel {
	imc := &v1alpha1.NatssChannel{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: v1alpha1.NatssChannelSpec{},
	}
	for _, opt := range imcopt {
		opt(imc)
	}
	imc.SetDefaults(context.Background())
	return imc
}

func WithInitNatssChannelConditions(imc *v1alpha1.NatssChannel) {
	imc.Status.InitializeConditions()
}

func WithNatssChannelDeleted(imc *v1alpha1.NatssChannel) {
	deleteTime := metav1.NewTime(time.Unix(1e9, 0))
	imc.ObjectMeta.SetDeletionTimestamp(&deleteTime)
}

func WithNatssChannelDeploymentNotReady(reason, message string) NatssChannelOption {
	return func(imc *v1alpha1.NatssChannel) {
		imc.Status.MarkDispatcherFailed(reason, message)
	}
}

func WithNatssChannelDeploymentReady() NatssChannelOption {
	return func(imc *v1alpha1.NatssChannel) {
		imc.Status.PropagateDispatcherStatus(&appsv1.DeploymentStatus{Conditions: []appsv1.DeploymentCondition{{Type: appsv1.DeploymentAvailable, Status: corev1.ConditionTrue}}})
	}
}

func WithNatssChannelServicetNotReady(reason, message string) NatssChannelOption {
	return func(imc *v1alpha1.NatssChannel) {
		imc.Status.MarkServiceFailed(reason, message)
	}
}

func WithNatssChannelServiceReady() NatssChannelOption {
	return func(imc *v1alpha1.NatssChannel) {
		imc.Status.MarkServiceTrue()
	}
}

func WithNatssChannelChannelServicetNotReady(reason, message string) NatssChannelOption {
	return func(imc *v1alpha1.NatssChannel) {
		imc.Status.MarkChannelServiceFailed(reason, message)
	}
}

func WithNatssChannelChannelServiceReady() NatssChannelOption {
	return func(imc *v1alpha1.NatssChannel) {
		imc.Status.MarkChannelServiceTrue()
	}
}

func WithNatssChannelEndpointsNotReady(reason, message string) NatssChannelOption {
	return func(imc *v1alpha1.NatssChannel) {
		imc.Status.MarkEndpointsFailed(reason, message)
	}
}

func WithNatssChannelEndpointsReady() NatssChannelOption {
	return func(imc *v1alpha1.NatssChannel) {
		imc.Status.MarkEndpointsTrue()
	}
}

func WithNatssChannelAddress(a string) NatssChannelOption {
	return func(imc *v1alpha1.NatssChannel) {
		imc.Status.SetAddress(&apis.URL{
			Scheme: "http",
			Host:   a,
		})
	}
}
