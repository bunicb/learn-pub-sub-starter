package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	const rabbitConnString = "amqp://guest:guest@docker:5672/"

	rmq, err := amqp.Dial(rabbitConnString)
	if err != nil {
		log.Fatalf("could not connect to RabbitMQ: %v", err)
	}
	defer rmq.Close()
	fmt.Println("Peril game server connected to RabbitMQ!")
	ch, err := rmq.Channel()
	if err != nil {
		log.Fatalf("could not create a channel: %v", err)
	}

	if err := pubsub.PublishJSON(ch,
		string(routing.ExchangePerilDirect),
		string(routing.PauseKey),
		routing.PlayingState{
			IsPaused: true,
		}); err != nil {
		log.Fatalf("could not publish message: %v", err)
	}

	// wait for ctrl+c
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan
	fmt.Println("RabbitMQ connection closed.")
}
