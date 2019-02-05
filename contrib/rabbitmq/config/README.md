# RabbitMQ Channels

Deployment steps:

1. Setup [Knative Eventing](../../../DEVELOPMENT.md)

2. If not done already, install a [RabbitMQ](https://www.rabbitmq.com/) cluster.

3. Now that RabbitMQ is installed, you need to configure the
   `amqp_url` value in the `rabbitmq-channel-controller-config` ConfigMap,
   located inside the `contrib/rabbitmq/config/rabbitmq.yaml` file.
   
4. Apply the 'RabbitMQ' ClusterChannelProvisioner, Controller, and Dispatcher:

   ```
   ko apply -f contrib/rabbitmq/config/kafka.yaml
   ```

1. Create Channels that reference the 'rabbitmq' ClusterChannelProvisioner.

   ```yaml
   apiVersion: eventing.knative.dev/v1alpha1
   kind: Channel
   metadata:
     name: my-rabbitmq-channel
   spec:
     provisioner:
       apiVersion: eventing.knative.dev/v1alpha1
       kind: ClusterChannelProvisioner
       name: rabbitmq
   ```