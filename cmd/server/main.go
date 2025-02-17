package main

import (
	"log"
	"os"
	"os/signal"

	amqp "github.com/rabbitmq/amqp091-go"
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
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	sig := <-signalChan
	log.Printf("Signal was recieved: %s\n", sig.String())
	con.Close()

	os.Exit(0)
}
