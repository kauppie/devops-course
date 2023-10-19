package main

import (
	"context"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Publisher struct {
	channel *amqp.Channel
	topic   string
}

func NewPublisher(conn *amqp.Connection, topic string) (*Publisher, error) {
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

	_, err = ch.QueueDeclare(
		topic,
		false,
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
		topic:   topic,
	}, nil
}

func (p *Publisher) Close() error {
	if err := p.channel.Close(); err != nil {
		return err
	}
	return nil
}

func (p *Publisher) Publish(body string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return p.channel.PublishWithContext(ctx,
		p.topic,
		"#",
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(body),
		})
}
