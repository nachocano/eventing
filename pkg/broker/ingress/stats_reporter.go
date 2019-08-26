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

package ingress

import (
	"context"
	"fmt"
	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
	utils "knative.dev/eventing/pkg/broker"
	"knative.dev/eventing/pkg/metrics/metricskey"
	"knative.dev/pkg/metrics"
	"time"
)

var (
	// eventCountM is a counter which records the number of events received
	// by the Broker.
	eventCountM = stats.Int64(
		"event_count",
		"Number of events received by a Broker",
		stats.UnitDimensionless,
	)

	// dispatchTimeInMsecM records the time spent dispatching an event to
	// a Trigger, in milliseconds.
	dispatchTimeInMsecM = stats.Float64(
		"dispatch_latencies",
		"The time spent dispatching an event to a Trigger",
		stats.UnitMilliseconds,
	)
)

type ReportArgs struct {
	ns        string
	broker    string
	eventType string
}

// StatsReporter defines the interface for sending ingress metrics.
type StatsReporter interface {
	ReportEventCount(args *ReportArgs, err error) error
	ReportDispatchTime(args *ReportArgs, err error, d time.Duration) error
}

// Reporter holds cached metric objects to report ingress metrics.
type Reporter struct {
	initialized     bool
	namespaceTagKey tag.Key
	brokerTagKey    tag.Key
	eventTypeKey    tag.Key
	// TODO add support for EventSource
	resultKey tag.Key
}

// NewStatsReporter creates a reporter that collects and reports ingress metrics.
func NewStatsReporter() (*Reporter, error) {
	var r = &Reporter{}

	// Create the tag keys that will be used to add tags to our measurements.
	nsTag, err := tag.NewKey(metricskey.NamespaceName)
	if err != nil {
		return nil, err
	}
	r.namespaceTagKey = nsTag
	brokerTag, err := tag.NewKey(metricskey.BrokerName)
	if err != nil {
		return nil, err
	}
	r.brokerTagKey = brokerTag
	eventTypeTag, err := tag.NewKey(metricskey.EventType)
	if err != nil {
		return nil, err
	}
	r.eventTypeKey = eventTypeTag
	resultTag, err := tag.NewKey(metricskey.Result)
	if err != nil {
		return nil, err
	}
	r.resultKey = resultTag

	// Create view to see our measurements.
	err = view.Register(
		&view.View{
			Description: eventCountM.Description(),
			Measure:     eventCountM,
			Aggregation: view.Sum(),
			TagKeys:     []tag.Key{r.namespaceTagKey, r.brokerTagKey, r.eventTypeKey, r.resultKey},
		},
		&view.View{
			Description: dispatchTimeInMsecM.Description(),
			Measure:     dispatchTimeInMsecM,
			Aggregation: view.Distribution(utils.Buckets125(1, 100)...), // 1, 2, 5, 10, 20, 50, 100
			TagKeys:     []tag.Key{r.namespaceTagKey, r.brokerTagKey, r.eventTypeKey, r.resultKey},
		},
	)
	if err != nil {
		return nil, err
	}

	r.initialized = true
	return r, nil
}

// ReportEventCount captures the event count.
func (r *Reporter) ReportEventCount(args *ReportArgs, err error) error {
	if !r.initialized {
		return fmt.Errorf("StatsReporter is not initialized yet")
	}

	ctx, err := tag.New(
		context.Background(),
		tag.Insert(r.namespaceTagKey, args.ns),
		tag.Insert(r.brokerTagKey, args.broker),
		tag.Insert(r.eventTypeKey, args.eventType),
		tag.Insert(r.resultKey, utils.Result(err)))
	if err != nil {
		return err
	}

	metrics.Record(ctx, eventCountM.M(1))
	return nil
}

// ReportDispatchTime captures dispatch times.
func (r *Reporter) ReportDispatchTime(args *ReportArgs, err error, d time.Duration) error {
	if !r.initialized {
		return fmt.Errorf("StatsReporter is not initialized yet")
	}

	ctx, err := tag.New(
		context.Background(),
		tag.Insert(r.namespaceTagKey, args.ns),
		tag.Insert(r.brokerTagKey, args.broker),
		tag.Insert(r.eventTypeKey, args.eventType),
		tag.Insert(r.resultKey, utils.Result(err)))
	if err != nil {
		return err
	}

	// convert time.Duration in nanoseconds to milliseconds.
	metrics.Record(ctx, dispatchTimeInMsecM.M(float64(d/time.Millisecond)))
	return nil
}
