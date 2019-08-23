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
	"context"
	"fmt"
	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
	utils "knative.dev/eventing/pkg/broker"
	"knative.dev/eventing/pkg/metrics/metricskey"
	"knative.dev/pkg/metrics"
	"strconv"
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
	triggerTypeKey       tag.Key
	triggerSourceKey     tag.Key
	responseCodeKey      tag.Key
	responseCodeClassKey tag.Key
	filterResultKey      tag.Key
}

// NewStatsReporter creates a reporter that collects and reports filter metrics.
func NewStatsReporter() (*Reporter, error) {
	var r = &Reporter{}

	// Create the tag keys that will be used to add tags to our measurements.
	nsTag, err := tag.NewKey(metricskey.NamespaceName)
	if err != nil {
		return nil, err
	}
	r.namespaceTagKey = nsTag
	triggerTag, err := tag.NewKey(metricskey.TriggerName)
	if err != nil {
		return nil, err
	}
	r.triggerTagKey = triggerTag
	brokerTag, err := tag.NewKey(metricskey.BrokerName)
	if err != nil {
		return nil, err
	}
	r.brokerTagKey = brokerTag
	triggerTypeTag, err := tag.NewKey(metricskey.TriggerType)
	if err != nil {
		return nil, err
	}
	r.triggerTypeKey = triggerTypeTag
	triggerSourceKey, err := tag.NewKey(metricskey.TriggerSource)
	if err != nil {
		return nil, err
	}
	r.triggerSourceKey = triggerSourceKey
	filterResultTag, err := tag.NewKey(metricskey.FilterResult)
	if err != nil {
		return nil, err
	}
	r.filterResultKey = filterResultTag
	responseCodeTag, err := tag.NewKey(metricskey.ResponseCode)
	if err != nil {
		return nil, err
	}
	r.responseCodeKey = responseCodeTag
	responseCodeClassTag, err := tag.NewKey(metricskey.ResponseCodeClass)
	if err != nil {
		return nil, err
	}
	r.responseCodeClassKey = responseCodeClassTag

	// Create view to see our measurements.
	err = view.Register(
		&view.View{
			Description: eventCountM.Description(),
			Measure:     eventCountM,
			Aggregation: view.Count(),
			TagKeys:     []tag.Key{r.namespaceTagKey, r.triggerTagKey, r.brokerTagKey, r.triggerTypeKey, r.triggerSourceKey, r.responseCodeKey, r.responseCodeClassKey},
		},
		&view.View{
			Description: dispatchTimeInMsecM.Description(),
			Measure:     dispatchTimeInMsecM,
			Aggregation: view.Distribution(utils.Buckets125(1, 100)...), // 1, 2, 5, 10, 20, 50, 100
			TagKeys:     []tag.Key{r.namespaceTagKey, r.triggerTagKey, r.brokerTagKey, r.triggerTypeKey, r.triggerSourceKey, r.responseCodeKey, r.responseCodeClassKey},
		},
		&view.View{
			Description: filterTimeInMsecM.Description(),
			Measure:     filterTimeInMsecM,
			Aggregation: view.Distribution(utils.Buckets125(0.1, 10)...), // 0.1, 0.2, 0.5, 1, 2, 5, 10
			TagKeys:     []tag.Key{r.namespaceTagKey, r.triggerTypeKey, r.brokerTagKey, r.triggerTypeKey, r.triggerSourceKey, r.filterResultKey},
		},
		&view.View{
			Description: deliveryTimeInMsecM.Description(),
			Measure:     deliveryTimeInMsecM,
			Aggregation: view.Distribution(utils.Buckets125(1, 100)...), // 1, 2, 5, 10, 20, 50, 100
			TagKeys:     []tag.Key{r.namespaceTagKey, r.triggerTypeKey, r.brokerTagKey, r.triggerTypeKey, r.triggerSourceKey, r.responseCodeKey, r.responseCodeClassKey},
		},
	)
	if err != nil {
		return nil, err
	}

	r.initialized = true
	return r, nil
}

// ReportEventCount captures the event count.
func (r *Reporter) ReportEventCount(context context.Context, ns, trigger, broker, triggerType, triggerSource string, responseCode int) error {
	if !r.initialized {
		return fmt.Errorf("StatsReporter is not initialized yet")
	}

	// Note that service names can be an empty string, so it needs a special treatment.
	ctx, err := tag.New(
		context,
		tag.Insert(r.namespaceTagKey, ns),
		tag.Insert(r.triggerTagKey, trigger),
		tag.Insert(r.brokerTagKey, broker),
		tag.Insert(r.triggerTypeKey, triggerType),
		tag.Insert(r.triggerSourceKey, triggerSource),
		tag.Insert(r.responseCodeKey, strconv.Itoa(responseCode)),
		tag.Insert(r.responseCodeClassKey, utils.ResponseCodeClass(responseCode)))
	if err != nil {
		return err
	}

	metrics.Record(ctx, eventCountM.M(1))
	return nil
}
