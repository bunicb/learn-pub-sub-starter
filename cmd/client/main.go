package main

import (
	"fmt"
	"log"
	"strconv"

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

	ch, queue, err := pubsub.DeclareAndBind(
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

	_, warQueue, err := pubsub.DeclareAndBind(
		rmq,
		routing.ExchangePerilTopic,
		routing.WarRecognitionsPrefix,
		routing.WarRecognitionsPrefix+".*",
		pubsub.SimpleQueueDurable,
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Queue %v declared and bound!\n", warQueue.Name)

	gameState := gamelogic.NewGameState(user)

	if err := pubsub.SubscribeJSON(
		rmq,
		routing.ExchangePerilDirect,
		queue.Name,
		routing.PauseKey,
		pubsub.SimpleQueueTransient,
		handlerPause(gameState),
	); err != nil {
		log.Fatal(err)
	}

	if err := pubsub.SubscribeJSON(
		rmq,
		routing.ExchangePerilTopic,
		routing.ArmyMovesPrefix+"."+user,
		routing.ArmyMovesPrefix+".*",
		pubsub.SimpleQueueTransient,
		handlerMove(gameState, ch),
	); err != nil {
		log.Fatal(err)
	}

	if err := pubsub.SubscribeJSON(
		rmq,
		routing.ExchangePerilTopic,
		routing.WarRecognitionsPrefix,
		routing.WarRecognitionsPrefix+".*",
		pubsub.SimpleQueueDurable,
		handlerWarMsg(gameState, ch),
	); err != nil {
		log.Fatal(err)
	}

	for {
		input := gamelogic.GetInput()
		if len(input) == 0 {
			continue
		}
		switch input[0] {
		case "spawn":
			if err := gameState.CommandSpawn(input); err != nil {
				fmt.Printf("There was a problem spawning. Error: %v\n", err)
				continue
			}
		case "move":
			move, err := gameState.CommandMove(input)
			if err != nil {
				fmt.Printf("Failed to move unit %s to %s. Error: %v\n", input[2], input[1], err)
				continue
			}
			if err := pubsub.PublishJSON(ch,
				string(routing.ExchangePerilTopic),
				routing.ArmyMovesPrefix+"."+user,
				move); err != nil {
				log.Fatalf("could not publish message: %v", err)
			}
			fmt.Println("Move published successfully")
		case "status":
			gameState.CommandStatus()
		case "help":
			gamelogic.PrintClientHelp()
		case "spam":
			if len(input) < 2 {
				fmt.Println("usage: spam <n>")
				continue
			}
			num, err := strconv.Atoi(input[1])
			if err != nil {
				fmt.Println("Please provide a number of messages as the second argument")
				continue
			}
			for range num {
				msg := gamelogic.GetMaliciousLog()
				if err := PublishGameLog(ch, msg, user); err != nil {
					fmt.Println("Couldn't publish a spam log")
				}
			}
			fmt.Printf("Published %v malicious logs\n", num)
		case "quit":
			gamelogic.PrintQuit()
			return
		default:
			fmt.Println("Unknown command.")
		}
	}
}
