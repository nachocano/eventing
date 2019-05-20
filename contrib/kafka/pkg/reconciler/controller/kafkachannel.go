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

package controller

import (
	"context"
	"reflect"
	"time"

	"github.com/knative/eventing/contrib/kafka/pkg/apis/messaging/v1alpha1"
	"github.com/knative/eventing/contrib/kafka/pkg/client/clientset/versioned"
	messaginginformers "github.com/knative/eventing/contrib/kafka/pkg/client/informers/externalversions/messaging/v1alpha1"
	listers "github.com/knative/eventing/contrib/kafka/pkg/client/listers/messaging/v1alpha1"
	"github.com/knative/eventing/pkg/logging"
	"github.com/knative/eventing/pkg/reconciler"
	"github.com/knative/pkg/controller"
	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	appsv1informers "k8s.io/client-go/informers/apps/v1"
	corev1informers "k8s.io/client-go/informers/core/v1"
	appsv1listers "k8s.io/client-go/listers/apps/v1"
	corev1listers "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
)

const (
	// ReconcilerName is the name of the reconciler.
	ReconcilerName = "KafkaChannels"

	// controllerAgentName is the string used by this controller to identify
	// itself when creating events.
	controllerAgentName = "kafka-controller"

	finalizerName = controllerAgentName

	// Name of the corev1.Events emitted from the reconciliation process.
	channelReconciled         = "ChannelReconciled"
	channelReconcileFailed    = "ChannelReconcileFailed"
	channelUpdateStatusFailed = "ChannelUpdateStatusFailed"
)

// Reconciler reconciles Kafka Channels.
type Reconciler struct {
	*reconciler.Base

	dispatcherNamespace      string
	dispatcherDeploymentName string
	dispatcherServiceName    string

	eventingClientSet    *versioned.Clientset
	kafkachannelLister   listers.KafkaChannelLister
	kafkachannelInformer cache.SharedIndexInformer
	deploymentLister     appsv1listers.DeploymentLister
	serviceLister        corev1listers.ServiceLister
	endpointsLister      corev1listers.EndpointsLister
	impl                 *controller.Impl
}

var (
	deploymentGVK = appsv1.SchemeGroupVersion.WithKind("Deployment")
	serviceGVK    = corev1.SchemeGroupVersion.WithKind("Service")
)

// Check that our Reconciler implements controller.Reconciler.
var _ controller.Reconciler = (*Reconciler)(nil)

// Check that our Reconciler implements cache.ResourceEventHandler
var _ cache.ResourceEventHandler = (*Reconciler)(nil)

// NewController initializes the controller and is called by the generated code.
// Registers event handlers to enqueue events.
func NewController(
	opt reconciler.Options,
	eventingClientSet *versioned.Clientset,
	dispatcherNamespace string,
	dispatcherDeploymentName string,
	dispatcherServiceName string,
	kafkachannelInformer messaginginformers.KafkaChannelInformer,
	deploymentInformer appsv1informers.DeploymentInformer,
	serviceInformer corev1informers.ServiceInformer,
	endpointsInformer corev1informers.EndpointsInformer,
) *controller.Impl {

	r := &Reconciler{
		Base:                     reconciler.NewBase(opt, controllerAgentName),
		dispatcherNamespace:      dispatcherNamespace,
		dispatcherDeploymentName: dispatcherDeploymentName,
		dispatcherServiceName:    dispatcherServiceName,
		eventingClientSet:        eventingClientSet,
		kafkachannelLister:       kafkachannelInformer.Lister(),
		kafkachannelInformer:     kafkachannelInformer.Informer(),
		deploymentLister:         deploymentInformer.Lister(),
		serviceLister:            serviceInformer.Lister(),
		endpointsLister:          endpointsInformer.Lister(),
	}
	r.impl = controller.NewImpl(r, r.Logger, ReconcilerName, reconciler.MustNewStatsReporter(ReconcilerName, r.Logger))

	r.Logger.Info("Setting up event handlers")
	kafkachannelInformer.Informer().AddEventHandler(reconciler.Handler(r.impl.Enqueue))

	// Set up watches for dispatcher resources we care about, since any changes to these
	// resources will affect our Channels. So, set up a watch here, that will cause
	// a global Resync for all the channels to take stock of their health when these change.
	deploymentInformer.Informer().AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: controller.FilterWithNameAndNamespace(dispatcherNamespace, dispatcherDeploymentName),
		Handler:    r,
	})
	serviceInformer.Informer().AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: controller.FilterWithNameAndNamespace(dispatcherNamespace, dispatcherServiceName),
		Handler:    r,
	})
	endpointsInformer.Informer().AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: controller.FilterWithNameAndNamespace(dispatcherNamespace, dispatcherServiceName),
		Handler:    r,
	})
	return r.impl
}

// cache.ResourceEventHandler implementation.
// These 3 functions just cause a Global Resync of the channels, because any changes here
// should be reflected onto the channels.
func (r *Reconciler) OnAdd(obj interface{}) {
	r.impl.GlobalResync(r.kafkachannelInformer)
}

func (r *Reconciler) OnUpdate(old, new interface{}) {
	r.impl.GlobalResync(r.kafkachannelInformer)
}

func (r *Reconciler) OnDelete(obj interface{}) {
	r.impl.GlobalResync(r.kafkachannelInformer)
}

// Reconcile compares the actual state with the desired, and attempts to
// converge the two. It then updates the Status block of the KafkaChannel resource
// with the current status of the resource.
func (r *Reconciler) Reconcile(ctx context.Context, key string) error {
	// Convert the namespace/name string into a distinct namespace and name.
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		logging.FromContext(ctx).Error("invalid resource key")
		return nil
	}

	// Get the KafkaChannel resource with this namespace/name.
	original, err := r.kafkachannelLister.KafkaChannels(namespace).Get(name)
	if apierrs.IsNotFound(err) {
		// The resource may no longer exist, in which case we stop processing.
		logging.FromContext(ctx).Error("KafkaChannel key in work queue no longer exists")
		return nil
	} else if err != nil {
		return err
	}

	// Don't modify the informers copy.
	channel := original.DeepCopy()

	// Reconcile this copy of the KafkaChannel and then write back any status updates regardless of
	// whether the reconcile error out.
	reconcileErr := r.reconcile(ctx, channel)
	if reconcileErr != nil {
		logging.FromContext(ctx).Error("Error reconciling KafkaChannel", zap.Error(reconcileErr))
		r.Recorder.Eventf(channel, corev1.EventTypeWarning, channelReconcileFailed, "KafkaChannel reconciliation failed: %v", reconcileErr)
	} else {
		logging.FromContext(ctx).Debug("KafkaChannel reconciled")
		r.Recorder.Event(channel, corev1.EventTypeNormal, channelReconciled, "KafkaChannel reconciled")
	}

	if _, updateStatusErr := r.updateStatus(ctx, channel); updateStatusErr != nil {
		logging.FromContext(ctx).Error("Failed to update GoogleCloudPubSubChannel status", zap.Error(updateStatusErr))
		r.Recorder.Eventf(channel, corev1.EventTypeWarning, channelUpdateStatusFailed, "Failed to update KafkaChannel's status: %v", updateStatusErr)
		return updateStatusErr
	}

	// Requeue if the resource is not ready
	return reconcileErr
}

func (r *Reconciler) reconcile(ctx context.Context, kc *v1alpha1.KafkaChannel) error {
	kc.Status.InitializeConditions()

	if kc.DeletionTimestamp != nil {
		// Everything is cleaned up by the garbage collector.
		return nil
	}

	// We reconcile the status of the Channel by looking at:
	// 1. Dispatcher Deployment for it's readiness.
	// 2. Dispatcher k8s Service for it's existence.
	// 3. Dispatcher endpoints to ensure that there's something backing the Service.
	// 4. k8s service representing the channel that will use ExternalName to point to the Dispatcher k8s service

	// Get the Dispatcher Deployment and propagate the status to the Channel
	//d, err := r.deploymentLister.Deployments(r.dispatcherNamespace).Get(r.dispatcherDeploymentName)
	//if err != nil {
	//	if apierrs.IsNotFound(err) {
	//		imc.Status.MarkDispatcherFailed("DispatcherDeploymentDoesNotExist", "Dispatcher Deployment does not exist")
	//	} else {
	//		logging.FromContext(ctx).Error("Unable to get the dispatcher Deployment", zap.Error(err))
	//		imc.Status.MarkDispatcherFailed("DispatcherDeploymentGetFailed", "Failed to get dispatcher Deployment")
	//	}
	//	return err
	//}
	//imc.Status.PropagateDispatcherStatus(&d.Status)
	//
	//// Get the Dispatcher Service and propagate the status to the Channel in case it does not exist.
	//// We don't do anything with the service because it's status contains nothing useful, so just do
	//// an existence check. Then below we check the endpoints targeting it.
	//_, err = r.serviceLister.Services(r.dispatcherNamespace).Get(r.dispatcherServiceName)
	//if err != nil {
	//	if apierrs.IsNotFound(err) {
	//		imc.Status.MarkServiceFailed("DispatcherServiceDoesNotExist", "Dispatcher Service does not exist")
	//	} else {
	//		logging.FromContext(ctx).Error("Unable to get the dispatcher service", zap.Error(err))
	//		imc.Status.MarkServiceFailed("DispatcherServiceGetFailed", "Failed to get dispatcher service")
	//	}
	//	return err
	//}
	//
	//imc.Status.MarkServiceTrue()
	//
	//// Get the Dispatcher Service Endpoints and propagate the status to the Channel
	//// endpoints has the same name as the service, so not a bug.
	//e, err := r.endpointsLister.Endpoints(r.dispatcherNamespace).Get(r.dispatcherServiceName)
	//if err != nil {
	//	if apierrs.IsNotFound(err) {
	//		imc.Status.MarkEndpointsFailed("DispatcherEndpointsDoesNotExist", "Dispatcher Endpoints does not exist")
	//	} else {
	//		logging.FromContext(ctx).Error("Unable to get the dispatcher endpoints", zap.Error(err))
	//		imc.Status.MarkEndpointsFailed("DispatcherEndpointsGetFailed", "Failed to get dispatcher endpoints")
	//	}
	//	return err
	//}
	//
	//if len(e.Subsets) == 0 {
	//	logging.FromContext(ctx).Error("No endpoints found for Dispatcher service", zap.Error(err))
	//	imc.Status.MarkEndpointsFailed("DispatcherEndpointsNotReady", "There are no endpoints ready for Dispatcher service")
	//	return errors.New("there are no endpoints ready for Dispatcher service")
	//}
	//
	//imc.Status.MarkEndpointsTrue()
	//
	//// Reconcile the k8s service representing the actual Channel. It points to the Dispatcher service via
	//// ExternalName
	//svc, err := r.reconcileChannelService(ctx, imc)
	//if err != nil {
	//	imc.Status.MarkChannelServiceFailed("ChannelServiceFailed", fmt.Sprintf("Channel Service failed: %s", err))
	//	return err
	//}
	//imc.Status.MarkChannelServiceTrue()
	//imc.Status.SetAddress(fmt.Sprintf("%s.%s.svc.%s", svc.Name, svc.Namespace, utils.GetClusterDomainName()))

	// Ok, so now the Dispatcher Deployment & Service have been created, we're golden since the
	// dispatcher watches the Channel and where it needs to dispatch events to.
	return nil
}

//func (r *Reconciler) reconcileChannelService(ctx context.Context, imc *v1alpha1.InMemoryChannel) (*corev1.Service, error) {
//	// Get the  Service and propagate the status to the Channel in case it does not exist.
//	// We don't do anything with the service because it's status contains nothing useful, so just do
//	// an existence check. Then below we check the endpoints targeting it.
//	// We may change this name later, so we have to ensure we use proper addressable when resolving these.
//	svc, err := r.serviceLister.Services(imc.Namespace).Get(fmt.Sprintf("%s-kn-channel", imc.Name))
//	if err != nil {
//		if apierrs.IsNotFound(err) {
//			svc, err = resources.NewK8sService(imc, resources.ExternalService(r.dispatcherNamespace, r.dispatcherServiceName))
//			if err != nil {
//				logging.FromContext(ctx).Error("failed to create the channel service object", zap.Error(err))
//				return nil, err
//			}
//			svc, err = r.KubeClientSet.CoreV1().Services(imc.Namespace).Create(svc)
//			if err != nil {
//				logging.FromContext(ctx).Error("failed to create the channel service", zap.Error(err))
//				return nil, err
//			}
//			return svc, nil
//		} else {
//			logging.FromContext(ctx).Error("Unable to get the channel service", zap.Error(err))
//		}
//		return nil, err
//	}
//
//	// Check to make sure that our IMC owns this service and if not, complain.
//	if !metav1.IsControlledBy(svc, imc) {
//		return nil, fmt.Errorf("inmemorychannel: %s/%s does not own Service: %q", imc.Namespace, imc.Name, svc.Name)
//	}
//	return svc, nil
//}

func (r *Reconciler) updateStatus(ctx context.Context, desired *v1alpha1.KafkaChannel) (*v1alpha1.KafkaChannel, error) {
	kc, err := r.kafkachannelLister.KafkaChannels(desired.Namespace).Get(desired.Name)
	if err != nil {
		return nil, err
	}

	if reflect.DeepEqual(kc.Status, desired.Status) {
		return kc, nil
	}

	becomesReady := desired.Status.IsReady() && !kc.Status.IsReady()

	// Don't modify the informers copy.
	existing := kc.DeepCopy()
	existing.Status = desired.Status

	new, err := r.eventingClientSet.MessagingV1alpha1().KafkaChannels(desired.Namespace).UpdateStatus(existing)
	if err == nil && becomesReady {
		duration := time.Since(new.ObjectMeta.CreationTimestamp.Time)
		r.Logger.Infof("KafkaChannel %q became ready after %v", kc.Name, duration)
		// TODO: stats
	}

	return new, err
}
