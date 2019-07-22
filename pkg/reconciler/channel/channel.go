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

package channel

import (
	"context"
	"fmt"
	"reflect"
	"time"

	corev1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/tools/cache"

	"github.com/knative/eventing/pkg/apis/messaging/v1alpha1"
	listers "github.com/knative/eventing/pkg/client/listers/messaging/v1alpha1"
	"github.com/knative/eventing/pkg/logging"
	"github.com/knative/eventing/pkg/reconciler"
	"go.uber.org/zap"
	"knative.dev/pkg/controller"
)

const (
	channelReadinessChanged   = "ChannelReadinessChanged"
	channelReconciled         = "ChannelReconciled"
	channelUpdateStatusFailed = "ChannelUpdateStatusFailed"
)

type Reconciler struct {
	*reconciler.Base

	// listers index properties about resources
	channelLister listers.ChannelLister
}

// Check that our Reconciler implements controller.Reconciler
var _ controller.Reconciler = (*Reconciler)(nil)

func (r *Reconciler) Reconcile(ctx context.Context, key string) error {
	// Convert the namespace/name string into a distinct namespace and name
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		logging.FromContext(ctx).Error("invalid resource key")
		return nil
	}

	// Get the Channel resource with this namespace/name
	original, err := r.channelLister.Channels(namespace).Get(name)
	if apierrs.IsNotFound(err) {
		// The resource may no longer exist, in which case we stop processing.
		logging.FromContext(ctx).Error("Channel key in work queue no longer exists")
		return nil
	} else if err != nil {
		return err
	}

	// Delete is a no-op.
	if original.DeletionTimestamp != nil {
		return nil
	}

	// Don't modify the informers copy
	channel := original.DeepCopy()

	// Reconcile this copy of the Channel and then write back any status
	// updates regardless of whether the reconcile error out.
	reconcileErr := r.reconcile(ctx, channel)
	if reconcileErr != nil {
		logging.FromContext(ctx).Error("Error reconciling Channel", zap.Error(reconcileErr))
	} else {
		logging.FromContext(ctx).Debug("Successfully reconciled Channel")
		r.Recorder.Eventf(channel, corev1.EventTypeNormal, channelReconciled, "Channel reconciled: %s", key)
	}

	if _, updateStatusErr := r.updateStatus(ctx, channel.DeepCopy()); updateStatusErr != nil {
		logging.FromContext(ctx).Warn("Error updating Channel status", zap.Error(updateStatusErr))
		r.Recorder.Eventf(channel, corev1.EventTypeWarning, channelUpdateStatusFailed, "Failed to update channel status: %s", key)
		return updateStatusErr
	}

	// Requeue if the resource is not ready:
	return reconcileErr
}

func (r *Reconciler) reconcile(ctx context.Context, ch *v1alpha1.Channel) error {

	// TODO reconcile the channel
	logging.FromContext(ctx).Sugar().Infof("Reconciling Channel: %s", ch.Name)
	return nil
}

func (r *Reconciler) updateStatus(ctx context.Context, desired *v1alpha1.Channel) (*v1alpha1.Channel, error) {
	channel, err := r.channelLister.Channels(desired.Namespace).Get(desired.Name)
	if err != nil {
		return nil, err
	}

	// If there's nothing to update, just return.
	if reflect.DeepEqual(channel.Status, desired.Status) {
		return channel, nil
	}

	becomesReady := desired.Status.IsReady() && !channel.Status.IsReady()

	// Don't modify the informers copy.
	existing := channel.DeepCopy()
	existing.Status = desired.Status

	c, err := r.EventingClientSet.MessagingV1alpha1().Channels(desired.Namespace).UpdateStatus(existing)

	if err == nil && becomesReady {
		duration := time.Since(c.ObjectMeta.CreationTimestamp.Time)
		logging.FromContext(ctx).Sugar().Infof("Channel %q became ready after %v", channel.Name, duration)
		r.Recorder.Event(channel, corev1.EventTypeNormal, channelReadinessChanged, fmt.Sprintf("Channel %q became ready", channel.Name))
		if err := r.StatsReporter.ReportReady("Channel", channel.Namespace, channel.Name, duration); err != nil {
			logging.FromContext(ctx).Sugar().Infof("failed to record ready for Channel, %v", err)
		}
	}

	return c, err
}
