package main

import (
	"fmt"
	"log"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
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

	_, queue, err := pubsub.DeclareAndBind(
		rmq,
		routing.ExchangePerilTopic,
		routing.GameLogSlug,
		routing.GameLogSlug+".*",
		pubsub.SimpleQueueDurable,
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Queue %v declared and bound!\n", queue.Name)

	gamelogic.PrintServerHelp()

	for {
		input := gamelogic.GetInput()
		if len(input) == 0 {
			continue
		}
		switch input[0] {
		case "pause":
			fmt.Println("Sending pause message.")
			if err := pubsub.PublishJSON(ch,
				string(routing.ExchangePerilDirect),
				string(routing.PauseKey),
				routing.PlayingState{
					IsPaused: true,
				}); err != nil {
				log.Fatalf("could not publish message: %v", err)
			}
		case "resume":
			fmt.Println("Sending resume message.")
			if err := pubsub.PublishJSON(ch,
				string(routing.ExchangePerilDirect),
				string(routing.PauseKey),
				routing.PlayingState{
					IsPaused: false,
				}); err != nil {
				log.Fatalf("could not publish message: %v", err)
			}
		case "quit":
			fmt.Println("Exiting server.")
			return
		default:
			fmt.Println("Unknown command.")
		}

	}

}
