package queue

import (
	"context"
	"encoding/json"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQ struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

type Message struct {
	Body []byte
	msg  *amqp.Delivery
}

func (m *Message) Ack() {
	m.msg.Ack(false)
}

func (m *Message) Nack(requeue bool) {
	m.msg.Nack(false, requeue)
}

func NewRabbitMQ(url string) (*RabbitMQ, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	return &RabbitMQ{
		conn:    conn,
		channel: ch,
	}, nil
}

func (r *RabbitMQ) Close() {
	if r.channel != nil {
		r.channel.Close()
	}
	if r.conn != nil {
		r.conn.Close()
	}
}

func (r *RabbitMQ) DeclareQueue(name string) error {
	_, err := r.channel.QueueDeclare(
		name,  // name
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	return err
}

func (r *RabbitMQ) Publish(queueName string, data interface{}) error {
	body, err := json.Marshal(data)
	if err != nil {
		return err
	}

	return r.channel.PublishWithContext(
		context.Background(),
		"",        // exchange
		queueName, // routing key
		false,     // mandatory
		false,     // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
}

func (r *RabbitMQ) Consume(queueName string) (<-chan Message, error) {
	msgs, err := r.channel.Consume(
		queueName, // queue
		"",        // consumer
		false,     // auto-ack
		false,     // exclusive
		false,     // no-local
		false,     // no-wait
		nil,       // args
	)
	if err != nil {
		return nil, err
	}

	messages := make(chan Message)
	go func() {
		for msg := range msgs {
			messages <- Message{
				Body: msg.Body,
				msg:  &msg,
			}
		}
		close(messages)
	}()

	return messages, nil
}
