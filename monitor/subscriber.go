package main

import (
	amqp "github.com/rabbitmq/amqp091-go"
)

type Subscriber struct {
	channel   *amqp.Channel
	queueName string
}

func NewSubscriber(conn *amqp.Connection) (*Subscriber, error) {
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
		LogsTopic,
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, err
	}

	err = channel.QueueBind(
		queue.Name,
		"#",
		LogsTopic,
		false,
		nil)
	if err != nil {
		return nil, err
	}

	return &Subscriber{
		channel:   channel,
		queueName: queue.Name,
	}, nil
}

func (s *Subscriber) Channel() (<-chan amqp.Delivery, error) {
	return s.channel.Consume(
		s.queueName,
		"monitor",
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
	return nil
}
