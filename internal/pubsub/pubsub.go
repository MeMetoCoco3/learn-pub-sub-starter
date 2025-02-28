package pubsub

import (
	"context"
	"encoding/json"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type (
	QueueType int
	AckType   int
)

const (
	DURABLE QueueType = iota
	TRANSIENT
)

const (
	Ack AckType = iota
	NackRequeue
	NackDiscard
)

func PublishJson[T any](ch *amqp.Channel, exchange, key string, val T) error {
	bytes, err := json.Marshal(val)
	if err != nil {
		return err
	}
	return ch.PublishWithContext(context.Background(), exchange, key, false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        bytes,
	})
}

func PublishGob[T any](ch *amqp.Channel, exchange, key string, val T) error {
	bytes, err := EncodeGob(val)
	if err != nil {
		return err
	}

	return ch.PublishWithContext(context.Background(), exchange, key, false, false, amqp.Publishing{
		ContentType: "application/gob",
		Body:        bytes,
	})
}

func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType QueueType,
	handler func(T) AckType,
) error {
	unmarshaller := func(data []byte) (T, error) {
		var v T
		err := json.Unmarshal(data, &v)
		return v, err
	}
	return subscribe(conn, exchange, queueName, key, queueType, handler, unmarshaller)
}

func SubscribeGob[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType QueueType,
	handler func(T) AckType,
) error {
	return subscribe(conn, exchange, queueName, key, queueType, handler, DecodeGob)
}

func DeclareAndBind(
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	simpleQueueType QueueType,
) (*amqp.Channel, amqp.Queue, error) {
	chann, err := conn.Channel()
	if err != nil {
		return nil, amqp.Queue{}, err
	}

	// Durable default
	durable := true
	autoDelete := false
	exclusive := false

	if simpleQueueType == TRANSIENT {
		durable = false
		autoDelete = true
		exclusive = true
	}
	queue, err := chann.QueueDeclare(
		queueName,
		durable,
		autoDelete,
		exclusive,
		false,
		amqp.Table{
			// Dice a Rabbit que ese es el exchange del deadletter
			"x-dead-letter-exchange": "peril_dlx",
		})
	if err != nil {
		return nil, amqp.Queue{}, err
	}

	chann.QueueBind(queueName, key, exchange, false, nil)

	return chann, queue, nil
}

func subscribe[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType QueueType,
	handler func(T) AckType,
	unmarshaller func([]byte) (T, error),
) error {
	// Check if exist
	chann, queue, err := DeclareAndBind(conn, exchange, queueName, key, queueType)
	if err != nil {
		return err
	}

	// We get a delivery chanel with a consumer name (if "" it will be autogenerated)
	deliveryChannel, err := chann.Consume(queue.Name, "", false, false, false, false, nil)
	if err != nil {
		return err
	}

	go func() {
		for msg := range deliveryChannel {

			val, err := unmarshaller(msg.Body)
			if err != nil {
				log.Printf("Error found: %s", err)
				continue
			}
			respose := handler(val)
			switch respose {
			case Ack:
				msg.Ack(false)
			case NackRequeue:
				msg.Nack(false, true)
			case NackDiscard:
				msg.Nack(false, false)
			}
		}
	}()
	return nil
}
