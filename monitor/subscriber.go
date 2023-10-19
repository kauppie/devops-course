package main

import (
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"
)

type Subscriber struct {
	channel   *amqp.Channel
	conn      *amqp.Connection
	queueName string
}

func NewSubscriber(addr string) (*Subscriber, error) {
	conn, err := amqp.Dial(addr)
	for err != nil {
		logrus.Warn("failed to connect; retrying in 2 seconds")
		conn, err = amqp.Dial(addr)

		<-time.After(2 * time.Second)
	}

	channel, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	err = channel.ExchangeDeclare(
		LogsTopic,
		"topic",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, err
	}

	queue, err := channel.QueueDeclare(
		"",
		false,
		false,
		true,
		false,
		nil,
	)
	if err != nil {
		return nil, err
	}

	err = channel.QueueBind(
		queue.Name,
		"",
		LogsTopic,
		false,
		nil)
	if err != nil {
		return nil, err
	}

	return &Subscriber{
		channel:   channel,
		conn:      conn,
		queueName: queue.Name,
	}, nil
}

func (s *Subscriber) Channel() (<-chan amqp.Delivery, error) {
	return s.channel.Consume(
		s.queueName,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
}

func (s *Subscriber) Close() error {
	if err := s.channel.Close(); err != nil {
		return err
	}
	if err := s.conn.Close(); err != nil {
		return err
	}
	return nil
}
