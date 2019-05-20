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

package v1alpha1

import (
	"context"
	. "github.com/knative/eventing/contrib/kafka/pkg/reconciler"
)

func (c *KafkaChannel) SetDefaults(ctx context.Context) {
	c.Spec.SetDefaults(ctx)
}

func (cs *KafkaChannelSpec) SetDefaults(ctx context.Context) {
	if cs.ConsumerMode == "" {
		cs.ConsumerMode = ConsumerModeMultiplexConsumerValue
	}
	if cs.NumPartitions == 0 {
		cs.NumPartitions = DefaultNumPartitions
	}
	if cs.ReplicationFactor == 0 {
		cs.ReplicationFactor = DefaultReplicationFactor
	}
}
