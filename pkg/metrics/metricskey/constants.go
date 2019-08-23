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

package metricskey

import "k8s.io/apimachinery/pkg/util/sets"

const (
	// KnativeTrigger is the Stackdriver resource type for Triggers.
	KnativeTrigger = "knative_trigger"

	// Project is the label for the project (e.g., GCP project ID).
	Project = "project_id"

	// Location is the label for the location (e.g. GCE zone) where the cluster is deployed.
	Location = "location"

	// ClusterName is the label for the immutable name of the cluster.
	ClusterName = "cluster_name"

	// NamespaceName is the label for the immutable name of the namespace where the resource type exists.
	NamespaceName = "namespace_name"

	// TriggerName is the label for the name of the trigger.
	TriggerName = "trigger_name"

	// TriggerType is the label for the type attribute filter of the Trigger.
	TriggerType = "trigger_type"

	// TriggerSource is the label for the source attribute filter of the Trigger.
	TriggerSource = "trigger_source"

	// Unknown is the default value if the field is unknown, e.g., the project will be unknown if Knative
	// is not running on GKE.
	Unknown = "unknown"
)

var (
	// KnativeTriggerLabels stores the set of resource labels for the resource type knative_trigger.
	KnativeTriggerLabels = sets.NewString(
		Project,
		Location,
		ClusterName,
		NamespaceName,
		TriggerName,
		TriggerType,
		TriggerSource,
	)

	// KnativeTriggerMetrics stores the set of metric types that are supported by the resource type knative_trigger.
	KnativeTriggerMetrics = sets.NewString(
		"knative.dev/eventing/trigger/event_count",
		"knative.dev/eventing/trigger/dispatch_latencies",
		"knative.dev/eventing/trigger/filter_latencies",
		// TODO event_latencies should be associated with Broker.
	)
)
