package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	ps "github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"

	routing "github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	log.Println("Starting Peril server...")

	/*==============================
	  CONNECTION
	==============================*/
	connectionString := "amqp://guest:guest@localhost:5672/"
	con, err := amqp.Dial(connectionString)
	if err != nil {
		log.Fatalf("Failed to connect: %s\n", err)
	}
	defer con.Close()

	log.Println("Successfully connected")

	chanel, err := con.Channel()
	if err != nil {
		log.Fatalf("Error creating new channel: %s", err)
	}

	/*==============================
	    EXCHANGE
		==============================*/
	err = chanel.ExchangeDeclare(
		routing.ExchangePerilDirect,
		"direct",
		true, false, false, false, nil)
	if err != nil {
		log.Fatalf("Error creating new channel: %s", err)
	}

	err = chanel.ExchangeDeclare(
		routing.ExchangePerilTopic,
		"topic",
		true, false, false, false, nil)
	if err != nil {
		log.Fatalf("Error creating new channel: %s", err)
	}

	/*==============================
	    QUEUES
		==============================*/
	_, err = chanel.QueueDeclare(
		"pause_test", true, false, false, false, nil)
	if err != nil {
		log.Fatalf("Error creating new channel: %s", err)
	}

	_, err = chanel.QueueDeclare(
		routing.GameLogSlug, true, false, false, false, amqp.Table{
			"x-dead-letter-exchange": "peril_dlx",
		})
	if err != nil {
		log.Fatalf("Error creating new channel: %s", err)
	}

	/*==============================
	    BINDING
		==============================*/
	chanel.QueueBind(
		"pause_test",
		routing.PauseKey,
		routing.ExchangePerilDirect,
		false, nil,
	)
	chanel.QueueBind(
		routing.GameLogSlug,
		"game_logs.*",
		routing.ExchangePerilTopic,
		false, nil,
	)

	/*==============================
	    DATA
		==============================*/
	data := routing.PlayingState{IsPaused: true}
	if err != nil {
		log.Fatalf("Error creating new channel: %s", err)
	}

	err = ps.PublishJson(chanel, routing.ExchangePerilDirect, routing.PauseKey, data)
	if err != nil {
		log.Fatalf("Error creating new channel: %s", err)
	}

	err = ps.SubscribeGob(
		con,
		routing.ExchangePerilTopic,
		routing.GameLogSlug,
		routing.GameLogSlug+".*",
		ps.DURABLE,
		func(logg routing.GameLog) ps.AckType {
			defer fmt.Print("> ")
			log.Printf("Log recieved: %s\n", logg.Message)
			gamelogic.WriteLog(logg)
			return ps.Ack
		},
	)
	if err != nil {
		log.Panic(err)
	}

	gamelogic.PrintServerHelp()
	gameOn := true
	for gameOn {
		input := gamelogic.GetInput()
		if len(input) == 0 {
			continue
		}

		for _, word := range input {
			switch word {
			case "pause":
				log.Println("Pausing the game")
				data := routing.PlayingState{IsPaused: true}
				ps.PublishJson(chanel, routing.ExchangePerilDirect, routing.PauseKey, data)

			case "resume":
				log.Println("Resuming the game")
				data := routing.PlayingState{IsPaused: false}
				ps.PublishJson(chanel, routing.ExchangePerilDirect, routing.PauseKey, data)

			case "quit":
				log.Println("Quiting the game")
				gameOn = false

			default:
				log.Printf("Not expected command: %s", word)

			}
		}

	}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	sig := <-signalChan
	log.Printf("Signal was recieved: %s\n", sig.String())
}
