/*
 * Copyright 2019 The Knative Authors
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

package v1alpha1

import (
	duckv1alpha1 "github.com/knative/pkg/apis/duck/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

var gcpsc = duckv1alpha1.NewLivingConditionSet(
	GoogleCloudPubSubChannelConditionDispatcherReady,
	GoogleCloudPubSubChannelConditionServiceReady,
	GoogleCloudPubSubChannelConditionEndpointsReady,
	GoogleCloudPubSubChannelConditionAddressable,
	GoogleCloudPubSubChannelConditionChannelServiceReady)

const (
	// GoogleCloudPubSubChannelConditionReady has status True when all subconditions below have been set to True.
	GoogleCloudPubSubChannelConditionReady = duckv1alpha1.ConditionReady

	// GoogleCloudPubSubChannelConditionDispatcherReady has status True when a Dispatcher deployment is ready
	// Keyed off appsv1.DeploymentAvailable, which means minimum available replicas required are up
	// and running for at least minReadySeconds.
	GoogleCloudPubSubChannelConditionDispatcherReady duckv1alpha1.ConditionType = "DispatcherReady"

	// GoogleCloudPubSubChannelConditionServiceReady has status True when a k8s Service is ready. This
	// basically just means it exists because there's no meaningful status in Service. See Endpoints
	// below.
	GoogleCloudPubSubChannelConditionServiceReady duckv1alpha1.ConditionType = "ServiceReady"

	// GoogleCloudPubSubChannelConditionEndpointsReady has status True when a k8s Service Endpoints are backed
	// by at least one endpoint.
	GoogleCloudPubSubChannelConditionEndpointsReady duckv1alpha1.ConditionType = "EndpointsReady"

	// GoogleCloudPubSubChannelConditionAddressable has status true when this GoogleCloudPubSubChannel meets
	// the Addressable contract and has a non-empty hostname.
	GoogleCloudPubSubChannelConditionAddressable duckv1alpha1.ConditionType = "Addressable"

	// GoogleCloudPubSubChannelConditionServiceReady has status True when a k8s Service representing the channel is ready.
	// Because this uses ExternalName, there are no endpoints to check.
	GoogleCloudPubSubChannelConditionChannelServiceReady duckv1alpha1.ConditionType = "ChannelServiceReady"
)

// GetCondition returns the condition currently associated with the given type, or nil.
func (cs *GoogleCloudPubSubChannelStatus) GetCondition(t duckv1alpha1.ConditionType) *duckv1alpha1.Condition {
	return gcpsc.Manage(cs).GetCondition(t)
}

// IsReady returns true if the resource is ready overall.
func (cs *GoogleCloudPubSubChannelStatus) IsReady() bool {
	return gcpsc.Manage(cs).IsHappy()
}

// InitializeConditions sets relevant unset conditions to Unknown state.
func (cs *GoogleCloudPubSubChannelStatus) InitializeConditions() {
	gcpsc.Manage(cs).InitializeConditions()
}

// TODO: Use the new beta duck types.
func (cs *GoogleCloudPubSubChannelStatus) SetAddress(hostname string) {
	cs.Address.Hostname = hostname
	if hostname != "" {
		gcpsc.Manage(cs).MarkTrue(GoogleCloudPubSubChannelConditionAddressable)
	} else {
		gcpsc.Manage(cs).MarkFalse(GoogleCloudPubSubChannelConditionAddressable, "EmptyHostname", "hostname is the empty string")
	}
}

func (cs *GoogleCloudPubSubChannelStatus) MarkDispatcherFailed(reason, messageFormat string, messageA ...interface{}) {
	gcpsc.Manage(cs).MarkFalse(GoogleCloudPubSubChannelConditionDispatcherReady, reason, messageFormat, messageA...)
}

// TODO: Unify this with the ones from Eventing. Say: Broker, Trigger.
func (cs *GoogleCloudPubSubChannelStatus) PropagateDispatcherStatus(ds *appsv1.DeploymentStatus) {
	for _, cond := range ds.Conditions {
		if cond.Type == appsv1.DeploymentAvailable {
			if cond.Status != corev1.ConditionTrue {
				cs.MarkDispatcherFailed("DispatcherNotReady", "Dispatcher Deployment is not ready: %s : %s", cond.Reason, cond.Message)
			} else {
				gcpsc.Manage(cs).MarkTrue(GoogleCloudPubSubChannelConditionDispatcherReady)
			}
		}
	}
}

func (cs *GoogleCloudPubSubChannelStatus) MarkServiceFailed(reason, messageFormat string, messageA ...interface{}) {
	gcpsc.Manage(cs).MarkFalse(GoogleCloudPubSubChannelConditionServiceReady, reason, messageFormat, messageA...)
}

func (cs *GoogleCloudPubSubChannelStatus) MarkServiceTrue() {
	gcpsc.Manage(cs).MarkTrue(GoogleCloudPubSubChannelConditionServiceReady)
}

func (cs *GoogleCloudPubSubChannelStatus) MarkChannelServiceFailed(reason, messageFormat string, messageA ...interface{}) {
	gcpsc.Manage(cs).MarkFalse(GoogleCloudPubSubChannelConditionChannelServiceReady, reason, messageFormat, messageA...)
}

func (cs *GoogleCloudPubSubChannelStatus) MarkChannelServiceTrue() {
	gcpsc.Manage(cs).MarkTrue(GoogleCloudPubSubChannelConditionChannelServiceReady)
}

func (cs *GoogleCloudPubSubChannelStatus) MarkEndpointsFailed(reason, messageFormat string, messageA ...interface{}) {
	gcpsc.Manage(cs).MarkFalse(GoogleCloudPubSubChannelConditionEndpointsReady, reason, messageFormat, messageA...)
}

func (cs *GoogleCloudPubSubChannelStatus) MarkEndpointsTrue() {
	gcpsc.Manage(cs).MarkTrue(GoogleCloudPubSubChannelConditionEndpointsReady)
}
