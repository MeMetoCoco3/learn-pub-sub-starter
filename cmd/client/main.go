package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	ps "github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	"github.com/rabbitmq/amqp091-go"
)

func main() {
	connectionString := "amqp://guest:guest@localhost:5672/"
	con, err := amqp091.Dial(connectionString)
	if err != nil {
		log.Fatal(err)
	}
	defer con.Close()

	userName, err := gamelogic.ClientWelcome()
	if err != nil {
		log.Fatal(err)
	}

	queueName := fmt.Sprintf("%s.%s", routing.PauseKey, userName)
	chann, queue, err := ps.DeclareAndBind(
		con,
		routing.ExchangePerilDirect,
		queueName,
		routing.PauseKey,
		ps.TRANSIENT)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Queue %v declared and bound!\n", queue.Name)

	game := gamelogic.NewGameState(userName)

	// We subscribe the client to the queues
	err = ps.SubscribeJSON(
		con,
		routing.ExchangePerilTopic,
		routing.ArmyMovesPrefix+"."+userName,
		routing.ArmyMovesPrefix+".*",
		ps.TRANSIENT,
		handlerMove(game))
	if err != nil {
		log.Panic(err)
	}

	err = ps.SubscribeJSON(
		con,
		routing.ExchangePerilDirect,
		queueName,
		routing.PauseKey,
		ps.TRANSIENT,
		handlerPause(game))
	if err != nil {
		log.Panic(err)
	}
	fmt.Println("Subscription by the client was successfull.")

	for {
		input := gamelogic.GetInput()
		if len(input) == 0 {
			continue
		}

		switch strings.ToLower(input[0]) {
		case "spawn":
			err := game.CommandSpawn(input)
			if err != nil {
				log.Println(err)
			}

		case "move":
			currentMove, err := game.CommandMove(input)
			if err != nil {
				log.Printf("Move unsuccessfull: %s\n", err)
				continue
			}

			err = ps.PublishJson(chann,
				routing.ExchangePerilTopic,
				routing.ArmyMovesPrefix+"."+userName,
				currentMove)
			if err != nil {
				log.Panic(err)
			}
			log.Println("Message was delivered successfully")

		case "status":
			game.CommandStatus()
		case "help":
			gamelogic.PrintClientHelp()

		case "quit":
			gamelogic.PrintQuit()
		default:
			log.Printf("Keyword not allowed: %s\n", input[0])
		}

	}
	/*
	   signalChan := make(chan os.Signal, 1)
	   signal.Notify(signalChan, os.Interrupt)
	   sig := <-signalChan
	   log.Println(sig.String())
	*/
}

func handlerPause(gs *gamelogic.GameState) func(routing.PlayingState) {
	return func(ps routing.PlayingState) {
		defer fmt.Print("> ")
		gs.HandlePause(ps)
	}
}

func handlerMove(gs *gamelogic.GameState) func(gamelogic.ArmyMove) {
	return func(move gamelogic.ArmyMove) {
		defer fmt.Println("> ")
		gs.HandleMove(move)
	}
}
