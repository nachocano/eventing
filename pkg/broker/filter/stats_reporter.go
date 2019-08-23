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

package filter

import (
	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
	utils "knative.dev/eventing/pkg/broker"
	"knative.dev/pkg/metrics/metricskey"
	"time"
)

var (
	// triggerEventCountM is a counter which records the number of events received
	// by a Trigger.
	eventCountM = stats.Int64(
		"event_count",
		"Number of events received by a Trigger",
		stats.UnitDimensionless,
	)

	// dispatchTimeInMsecM records the time spent dispatching an event for
	// a Trigger, in milliseconds.
	dispatchTimeInMsecM = stats.Float64(
		"dispatch_latencies",
		"The time spent dispatching an event to a Trigger",
		stats.UnitMilliseconds,
	)

	// filterTimeInMsecM records the time spent filtering a message for a
	// Trigger, in milliseconds.
	filterTimeInMsecM = stats.Float64(
		"filter_latencies",
		"The time spent filtering a message for a Trigger",
		stats.UnitMilliseconds,
	)

	// deliveryTimeInMsecM records the time spent between arrival at ingress
	// and delivery to the Trigger subscriber.
	deliveryTimeInMsecM = stats.Float64(
		"event_latencies",
		"The time spent dispatching an event from a Broker to a Trigger subscriber",
		stats.UnitMilliseconds,
	)

	// Tag keys must conform to the restrictions described in
	// go.opencensus.io/tag/validate.go. Currently those restrictions are:
	// - length between 1 and 255 inclusive
	// - characters are printable US-ASCII

	// TagFilterResult is a tag key referring to the observed result of a filter
	// operation.
	TagFilterResult = utils.MustNewTagKey("filter_result")

	// TagBroker is a tag key referring to the Broker name serviced by this
	// filter process.
	TagBroker = utils.MustNewTagKey("broker")

	// TagTrigger is a tag key referring to the Trigger name serviced by this
	// filter process.
	TagTrigger = utils.MustNewTagKey("trigger")

	// TagTrigger is a tag key referring to the Trigger type attribute filter.
	TagTriggerType = utils.MustNewTagKey("trigger_type")

	// TagEventType is a tag key referring to the CloudEvent type.
	TagEventType = utils.MustNewTagKey("event_type")
)

// StatsReporter defines the interface for sending filter metrics.
type StatsReporter interface {
	ReportEventCount(ns, broker, trigger, eventType string, responseCode int) error
	ReportDispatchTime(ns, broker, trigger, eventType string, responseCode int, d time.Duration) error
	ReportFilterTime(ns, broker, trigger, eventType string, responseCode int, d time.Duration) error
	ReportDeliveryTime(ns, broker, trigger, eventType string, responseCode int, d time.Duration) error
}

// Reporter holds cached metric objects to report filter metrics.
type Reporter struct {
	initialized          bool
	namespaceTagKey      tag.Key
	triggerTagKey        tag.Key
	brokerTagKey         tag.Key
	responseCodeKey      tag.Key
	responseCodeClassKey tag.Key
	filterResultKey      tag.Key
	triggerTypeKey       tag.Key
	eventTypeKey         tag.Key
}

// NewStatsReporter creates a reporter that collects and reports activator metrics
func NewStatsReporter() (*Reporter, error) {
	var r = &Reporter{}

	// Create the tag keys that will be used to add tags to our measurements.
	nsTag, err := tag.NewKey(metricskey.LabelNamespaceName)
	if err != nil {
		return nil, err
	}
	r.namespaceTagKey = nsTag
	serviceTag, err := tag.NewKey(metricskey.LabelServiceName)
	if err != nil {
		return nil, err
	}
	r.serviceTagKey = serviceTag
	configTag, err := tag.NewKey(metricskey.LabelConfigurationName)
	if err != nil {
		return nil, err
	}
	r.configTagKey = configTag
	revTag, err := tag.NewKey(metricskey.LabelRevisionName)
	if err != nil {
		return nil, err
	}
	r.revisionTagKey = revTag
	responseCodeTag, err := tag.NewKey("response_code")
	if err != nil {
		return nil, err
	}
	r.responseCodeKey = responseCodeTag
	responseCodeClassTag, err := tag.NewKey("response_code_class")
	if err != nil {
		return nil, err
	}
	r.responseCodeClassKey = responseCodeClassTag
	numTriesTag, err := tag.NewKey("num_tries")
	if err != nil {
		return nil, err
	}
	r.numTriesKey = numTriesTag
	// Create view to see our measurements.
	err = view.Register(
		&view.View{
			Description: "Concurrent requests that are routed to Activator",
			Measure:     requestConcurrencyM,
			Aggregation: view.LastValue(),
			TagKeys:     []tag.Key{r.namespaceTagKey, r.serviceTagKey, r.configTagKey, r.revisionTagKey},
		},
		&view.View{
			Description: "The number of requests that are routed to Activator",
			Measure:     requestCountM,
			Aggregation: view.Count(),
			TagKeys:     []tag.Key{r.namespaceTagKey, r.serviceTagKey, r.configTagKey, r.revisionTagKey, r.responseCodeKey, r.responseCodeClassKey, r.numTriesKey},
		},
		&view.View{
			Description: "The response time in millisecond",
			Measure:     responseTimeInMsecM,
			Aggregation: defaultLatencyDistribution,
			TagKeys:     []tag.Key{r.namespaceTagKey, r.serviceTagKey, r.configTagKey, r.revisionTagKey, r.responseCodeClassKey, r.responseCodeKey},
		},
	)
	if err != nil {
		return nil, err
	}

	r.initialized = true
	return r, nil
}

func init() {
	// Create views for exporting measurements. This returns an error if a
	// previously registered view has the same name with a different value.
	err := view.Register(
		&view.View{
			Name:        "trigger_events_total",
			Measure:     MeasureTriggerEventsTotal,
			Aggregation: view.Count(),
			TagKeys:     []tag.Key{TagResult, TagBroker, TagTrigger},
		},
		&view.View{
			Name:        "trigger_dispatch_time",
			Measure:     MeasureTriggerDispatchTime,
			Aggregation: view.Distribution(utils.Buckets125(1, 100)...), // 1, 2, 5, 10, 20, 50, 100,
			TagKeys:     []tag.Key{TagResult, TagBroker, TagTrigger},
		},
		&view.View{
			Name:        "trigger_filter_time",
			Measure:     MeasureTriggerFilterTime,
			Aggregation: view.Distribution(utils.Buckets125(0.1, 10)...), // 0.1, 0.2, 0.5, 1, 2, 5, 10
			TagKeys:     []tag.Key{TagResult, TagFilterResult, TagBroker, TagTrigger},
		},
		&view.View{
			Name:        "broker_to_function_delivery_time",
			Measure:     MeasureDeliveryTime,
			Aggregation: view.Distribution(utils.Buckets125(1, 100)...), // 1, 2, 5, 10, 20, 50, 100
			TagKeys:     []tag.Key{TagResult, TagBroker, TagTrigger},
		},
	)
	if err != nil {
		panic(err)
	}
}
