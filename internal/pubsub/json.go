package pubsub

import (
	"context"
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

type AckType int

const (
	Ack AckType = iota
	NackRequeue
	NackDiscard
)

func PublishJSON[T any](ch *amqp.Channel, exchange, key string, val T) error {
	data, err := json.Marshal(val)
	if err != nil {
		return err
	}
	return ch.PublishWithContext(context.Background(), exchange, key, false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        data,
	})
}

func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType, // an enum to represent "durable" or "transient"
	handler func(T) AckType,
) error {
	ch, queue, err := DeclareAndBind(
		conn,
		exchange,
		queueName,
		key,
		queueType,
	)
	if err != nil {
		return err
	}

	if err := ch.Qos(10, 0, false); err != nil {
		return err
	}

	deliveries, err := ch.Consume(queue.Name, "", false, false, false, false, nil)
	if err != nil {
		return err
	}

	go func() {
		defer ch.Close()
		for msg := range deliveries {
			var data T
			if err := json.Unmarshal(msg.Body, &data); err != nil {
				fmt.Printf("Couldn't unmarshal msg: %v\n", err)
				continue
			}
			switch handler(data) {
			case Ack:
				msg.Ack(false)
				fmt.Println("Ack")
			case NackRequeue:
				msg.Nack(false, true)
				fmt.Println("NackRequeue")
			case NackDiscard:
				msg.Nack(false, false)
				fmt.Println("NackDiscard")
			}
		}
	}()

	return nil
}
