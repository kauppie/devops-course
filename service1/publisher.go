package main

import (
	"context"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"
)

type Publisher struct {
	channel *amqp.Channel
	conn    *amqp.Connection
	topic   string
}

func NewPublisher(addr, topic string) (*Publisher, error) {
	conn, err := amqp.Dial(addr)
	for err != nil {
		logrus.Warn("failed to connect; retrying in 2 seconds")
		conn, err = amqp.Dial(addr)

		<-time.After(2 * time.Second)
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	err = ch.ExchangeDeclare(
		topic,
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

	return &Publisher{
		channel: ch,
		conn:    conn,
		topic:   topic,
	}, nil
}

func (p *Publisher) Close() error {
	if err := p.channel.Close(); err != nil {
		return err
	}
	if err := p.conn.Close(); err != nil {
		return err
	}
	return nil
}

func (p *Publisher) Publish(body string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return p.channel.PublishWithContext(ctx,
		p.topic,
		"",
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(body),
		})
}
