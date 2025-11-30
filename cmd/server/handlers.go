package main

import (
	"fmt"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
)

func handlerLogs() func(routing.GameLog) pubsub.AckType {
	return func(glog routing.GameLog) pubsub.AckType {
		defer fmt.Print("> ")
		err := gamelogic.WriteLog(glog)
		if err != nil {
			fmt.Println("Couldn't write log message")
			return pubsub.NackRequeue
		}
		return pubsub.Ack
	}
}
