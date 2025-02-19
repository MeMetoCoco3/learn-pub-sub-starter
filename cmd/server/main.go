package main

import (
	ps "github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"

	routing "github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
	"os"
	"os/signal"
)

func main() {
	log.Println("Starting Peril server...")

	connectionString := "amqp://guest:guest@localhost:5672/"
	con, err := amqp.Dial(connectionString)
	if err != nil {
		log.Fatal(err)
	}
	defer con.Close()

	log.Println("Successfully connected")

	chanel, err := con.Channel()
	if err != nil {
		log.Fatalf("Error creating new channel: %s", err)
	}
	data := routing.PlayingState{IsPaused: true}
	ps.PublishJson(chanel, routing.ExchangePerilDirect, routing.PauseKey, data)

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	sig := <-signalChan
	log.Printf("Signal was recieved: %s\n", sig.String())

	return
}
