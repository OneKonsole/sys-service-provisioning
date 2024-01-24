package utils

import (
	"context"
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitClient struct {
	// The connection to the used by the client
	conn *amqp.Connection
	// Channel is used to process / Send messages
	ch *amqp.Channel
}

// ConnectRabbitMQ connects to RabbitMQ and returns a connection
func ConnectRabbitMQ(username, password, host, vhost string) (*amqp.Connection, error) {
	return amqp.Dial(fmt.Sprintf(("amqp://%s:%s@%s/%s"), username, password, host, vhost))
}

// NewRabbitClient creates a new RabbitClient
func NewRabbitClient(conn *amqp.Connection) (RabbitClient, error) {
	ch, err := conn.Channel()
	if err != nil {
		return RabbitClient{}, err
	}

	if err := ch.Confirm(false); err != nil {
		return RabbitClient{}, err
	}

	return RabbitClient{
		conn: conn,
		ch:   ch,
	}, nil
}

// Qos sets the QoS of the channel
func (rc RabbitClient) Qos(prefetchCount, prefetchSize int, global bool) error {
	return rc.ch.Qos(prefetchCount, prefetchSize, global)
}

// CreateQueue creates a queue with the given name
func (rc RabbitClient) CreateQueue(queueName string, durable, autodelete bool) (amqp.Queue, error) {
	q, err := rc.ch.QueueDeclare(queueName, durable, autodelete, false, false, nil)
	if err != nil {
		log.Println("Error in the queue creation")
		return amqp.Queue{}, nil
	}
	return q, err
}

// CreateBinding creates a binding between a queue and an exchange
func (rc RabbitClient) CreateBinding(name, binding, exchange string) error {
	return rc.ch.QueueBind(name, binding, exchange, false, nil)
}

func (rc RabbitClient) Send(ctx context.Context, exchange, routingKey string, options amqp.Publishing) error {
	confirmation, err := rc.ch.PublishWithDeferredConfirmWithContext(
		ctx,
		exchange,
		routingKey,
		true,
		false,
		options,
	)

	if err != nil {
		return err
	}
	log.Println(confirmation.Wait())
	return nil
}

func (rc RabbitClient) Consume(queue, consumer string, autoAck bool) (<-chan amqp.Delivery, error) {
	return rc.ch.Consume(queue, consumer, autoAck, false, false, false, nil)
}

// Close closes the channel
func (rc RabbitClient) Close() error {
	return rc.ch.Close()
}
