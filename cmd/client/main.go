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
	fmt.Println("Starting Peril client...")
	const rabbitConnString = "amqp://guest:guest@docker:5672/"

	rmq, err := amqp.Dial(rabbitConnString)
	if err != nil {
		log.Fatalf("could not connect to RabbitMQ: %v", err)
	}
	defer rmq.Close()

	user, err := gamelogic.ClientWelcome()
	if err != nil {
		log.Fatal(err)
	}

	_, queue, err := pubsub.DeclareAndBind(
		rmq,
		routing.ExchangePerilDirect,
		fmt.Sprintf("%s.%s", routing.PauseKey, user),
		routing.PauseKey,
		pubsub.SimpleQueueTransient,
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Queue %v declared and bound!\n", queue.Name)

	gameState := gamelogic.NewGameState(user)

	for {
		input := gamelogic.GetInput()
		if len(input) == 0 {
			continue
		}
		switch input[0] {
		case "spawn":
			if err := gameState.CommandSpawn(input); err != nil {
				fmt.Printf("There was a problem spawning. Error: %v", err)
				continue
			}
		case "move":
			_, err := gameState.CommandMove(input)
			if err != nil {
				fmt.Printf("Failed to move unit %s to %s. Error: %v", input[2], input[1], err)
				continue
			}
		case "status":
			gameState.CommandStatus()
		case "help":
			gamelogic.PrintClientHelp()
		case "spam":
			fmt.Println("Spamming not allowed yet!")
		case "quit":
			gamelogic.PrintQuit()
			return
		default:
			fmt.Println("Unknown command.")
		}
	}
}
