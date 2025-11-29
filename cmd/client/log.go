package main

import (
	"time"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func PublishGameLog(ch *amqp.Channel, message, user string) error {
	log := routing.GameLog{
		CurrentTime: time.Now(),
		Message:     message,
		Username:    user,
	}
	err := pubsub.PublishGob(
		ch,
		routing.ExchangePerilTopic,
		routing.GameLogSlug+"."+user,
		log,
	)
	if err != nil {
		return err
	}
	return nil
}
