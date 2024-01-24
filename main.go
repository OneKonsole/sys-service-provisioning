package main

import (
	"context"
	"fmt"
	"os"

	"encoding/json"

	flags "github.com/jessevdk/go-flags"
	"github.com/onekonsole/sys-service-provisioning/internal/utils"
	"github.com/onekonsole/sys-service-provisioning/pkg/models"
	"golang.org/x/sync/errgroup"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	repository "github.com/onekonsole/sys-service-provisioning/internal/repositories"
	usecase "github.com/onekonsole/sys-service-provisioning/internal/usecases"
)

type Arguments struct {
	TypeKubernetesConnection string `short:"t" long:"type" description:"Type of Kubernetes connection" choice:"inCluster" choice:"kubeConfig" required:"true"`
	KubeConfigPath           string `short:"k" long:"kubeConfig" description:"Path to kubeconfig file"`
	Domain                   string `short:"d" long:"domain" description:"Domain name" required:"true"`
	ExposedIpAddress         string `short:"e" long:"exposedIpAddress" description:"Exposed IP adress" required:"true"`
	DataStore                string `short:"s" long:"datastore" description:"Datastore" required:"true"`
}

var arguments = Arguments{
	TypeKubernetesConnection: "kubeConfig",
	KubeConfigPath:           os.Getenv("HOME") + "/.kube/config",
	Domain:                   "",
	ExposedIpAddress:         "127.0.0.1",
}

var clientSet *kubernetes.Clientset
var concurencyLimit int = 3

func main() {
	_, err := flags.Parse(&arguments)
	if err != nil {
		fmt.Println("Error parsing flags: ", err)
		os.Exit(1)
	}

	// Connect to Kubernetes cluster and get clientSet
	switch arguments.TypeKubernetesConnection {
	case "inCluster":
		config, err := rest.InClusterConfig()
		if err != nil {
			fmt.Println("Error creating Kubernetes config: ", err)
			os.Exit(1)
		}
		clientSet, err = kubernetes.NewForConfig(config)
		if err != nil {
			fmt.Println("Error creating Kubernetes client: ", err)
			os.Exit(1)
		}
	case "kubeConfig":
		clientSet, err = utils.GetKubernetesClientsetFromFilePath(arguments.KubeConfigPath)
		if err != nil {
			fmt.Println("Error creating Kubernetes client: ", err)
			os.Exit(1)
		}
	}

	// Get rabbitMQ parameters from environment variables
	rabbitMQUser := os.Getenv("RABBITMQ_USER")
	rabbitMQPassword := os.Getenv("RABBITMQ_PASSWORD")
	rabbitMQHost := os.Getenv("RABBITMQ_HOST")
	rabbitMQQueue := os.Getenv("RABBITMQ_QUEUE")
	rabbitMQVHost := os.Getenv("RABBITMQ_VHOST")

	conn, err := utils.ConnectRabbitMQ(rabbitMQUser, rabbitMQPassword, rabbitMQHost, rabbitMQVHost)
	if err != nil {
		fmt.Println("Error connecting to RabbitMQ: ", err)
		panic(err)
	}
	defer conn.Close()

	rabbitClient, err := utils.NewRabbitClient(conn)
	if err != nil {
		fmt.Println("Error creating RabbitMQ client: ", err)
		panic(err)
	}
	defer rabbitClient.Close()

	// Define a QoS of X messages at a time similar to errgroup limit
	err = rabbitClient.Qos(concurencyLimit, 0, false)
	if err != nil {
		fmt.Println("Error setting QoS: ", err)
		panic(err)
	}

	// Consume messages from the queue "provisioning"
	messageBus, err := rabbitClient.Consume(rabbitMQQueue, "", false)
	if err != nil {
		fmt.Println("Error consuming messages from the queue: ", err)
		panic(err)
	}

	// Create a new context
	ctx := context.Background()
	// ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	// defer cancel()

	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(concurencyLimit)

	var blocking chan struct{}

	tenantRepository := repository.NewTenantKubernetesCluster(clientSet)
	tenantUseCase := usecase.NewTenantUseCase(tenantRepository, arguments.Domain, arguments.ExposedIpAddress)

	go func() {
		for message := range messageBus {
			// To avoid the shared variable problem
			msg := message
			g.Go(func() error {
				var order models.Order

				err := json.Unmarshal(msg.Body, &order)
				if err != nil {
					fmt.Println("Error while unmarshalling the message")
					msg.Nack(false, false)
					return nil
				}

				err = tenantUseCase.CreateTenant(ctx, order, order.UserID, arguments.DataStore)
				//TODO: Make an ack system
				if err != nil {
					fmt.Println("Error while provisioning the cluster " + order.ClusterName)
					// err := errors.New("Error while provisioning the cluster " + order.ClusterName)
					msg.Nack(false, false)
					return nil
				}
				err = msg.Ack(false) // Acknowledge the message
				if err != nil {
					fmt.Println("Error while acknowledging the message")
					msg.Nack(false, false)
					return nil
				}
				return nil
			})
		}
	}()

	<-blocking
}
