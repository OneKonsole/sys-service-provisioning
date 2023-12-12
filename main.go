package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"encoding/json"

	"github.com/onekonsole/sys-service-provisioning/cmd/provisioning"
	"github.com/onekonsole/sys-service-provisioning/internal"
	"github.com/onekonsole/sys-service-provisioning/pkg/models"
	"golang.org/x/sync/errgroup"
)

func main() {
	conn, err := internal.ConnectRabbitMQ("emdev", "secret", "localhost:5672", "provisioning")
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	rabbitClient, err := internal.NewRabbitClient(conn)
	if err != nil {
		panic(err)
	}
	defer rabbitClient.Close()

	// Consume messages from the queue "provisioning"
	messageBus, err := rabbitClient.Consume("provisioning", "", false)
	if err != nil {
		panic(err)
	}

	// Create a new context
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(5)

	var blocking chan struct{}

	go func() {
		for message := range messageBus {
			g.Go(func() error {
				var order models.Order

				err := json.Unmarshal(message.Body, &order)
				if err != nil {
					fmt.Println(err)
					return err
				}
				//TODO: Make an ack system
				isGood := provisioning.Run(order)
				if !isGood {
					fmt.Println("Error while provisioning the cluster " + order.ClusterName)
					err := errors.New("Error while provisioning the cluster " + order.ClusterName)
					return err
				}
				err = message.Ack(false) // Acknowledge the message
				if err != nil {
					fmt.Println(err)
					return err
				}
				return nil
			})
		}
	}()

	<-blocking
}
