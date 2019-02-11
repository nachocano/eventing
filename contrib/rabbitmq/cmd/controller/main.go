/*
 * Copyright 2018 The Knative Authors
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

package main

import (
	"flag"
	"fmt"

	"github.com/knative/eventing/pkg/provisioners"

	"github.com/knative/eventing/contrib/rabbitmq/pkg/channel"
	provisionerController "github.com/knative/eventing/contrib/rabbitmq/pkg/clusterchannelprovisioner"
	eventingv1alpha1 "github.com/knative/eventing/pkg/apis/eventing/v1alpha1"
	istiov1alpha3 "github.com/knative/pkg/apis/istio/v1alpha3"
	"github.com/knative/pkg/configmap"
	"github.com/knative/pkg/signals"
	"go.uber.org/zap"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

const (
	UrlConfigMapKey = "amqp_url"
)

// This is the main method for the RabbitMQ Channel controller. It reconciles the
// ClusterChannelProvisioner itself and Channels that use the 'rabbitmq' provisioner. It does not
// handle the anything at the data layer.
func main() {
	logConfig := provisioners.NewLoggingConfig()
	logger := provisioners.NewProvisionerLoggerFromConfig(logConfig)
	defer logger.Sync()
	logger = logger.With(
		zap.String("eventing.knative.dev/clusterChannelProvisioner", provisionerController.Name),
		zap.String("eventing.knative.dev/clusterChannelProvisionerComponent", "Controller"),
	)
	flag.Parse()

	mgr, err := manager.New(config.GetConfigOrDie(), manager.Options{})
	if err != nil {
		logger.Fatal("Error starting up.", zap.Error(err))
	}

	// Add custom types to this array to get them into the manager's scheme.
	eventingv1alpha1.AddToScheme(mgr.GetScheme())
	istiov1alpha3.AddToScheme(mgr.GetScheme())

	// The controllers for both the ClusterChannelProvisioner and the Channels created by that
	// ClusterChannelProvisioner run in this process.
	_, err = provisionerController.ProvideController(mgr, logger.Desugar())
	if err != nil {
		logger.Fatal("Unable to create Provisioner controller", zap.Error(err))
	}

	provisionerConfig, err := getProvisionerConfig()

	_, err = channel.ProvideController(provisionerConfig)(mgr, logger.Desugar())
	if err != nil {
		logger.Fatal("Unable to create Channel controller", zap.Error(err))
	}

	// set up signals so we handle the first shutdown signal gracefully
	stopCh := signals.SetupSignalHandler()
	// Start blocks forever.
	err = mgr.Start(stopCh)
	if err != nil {
		logger.Fatal("Manager.Start() returned an error", zap.Error(err))
	}
}

func getProvisionerConfig() (*provisionerController.RabbitMqProvisionerConfig, error) {
	configMap, err := configmap.Load("/etc/config-provisioner")
	if err != nil {
		return nil, fmt.Errorf("error loading provisioner configuration: %s", err)
	}
	if len(configMap) == 0 {
		return nil, fmt.Errorf("missing provisioner configuration")
	}
	config := &provisionerController.RabbitMqProvisionerConfig{}

	if url, ok := configMap[UrlConfigMapKey]; ok {
		config.Url = url
		return config, nil
	}
	return nil, fmt.Errorf("missing key %s in provisioner configuration", UrlConfigMapKey)
}
