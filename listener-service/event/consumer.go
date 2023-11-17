package event

import (
	"encoding/json"
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Consumer struct {
	conn      *amqp.Connection
	queueName string
}

func NewConsumer(conn *amqp.Connection) (Consumer, error) {
	consumer := Consumer{
		conn: conn,
	}

	if err := consumer.setup(); err != nil {
		return Consumer{}, err
	}

	return consumer, nil
}

func (consumer *Consumer) setup() error {
	channel, err := consumer.conn.Channel()
	if err != nil {
		return err
	}

	return declareExchange(channel)
}

type Payload struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

func (consumer *Consumer) Listen(topics []string) error {
	ch, err := consumer.conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	q, err := declareQueue(ch)
	if err != nil {
		return err
	}

	for _, topic := range topics {
		err = ch.QueueBind(
			q.Name,        // queue name
			topic,         // routing key
			EXCHANGE_NAME, // exchange name
			false,         // no-wait?
			nil,           // arguments
		)

		if err != nil {
			return err
		}
	}

	messageChannel, err := ch.Consume(
		q.Name, // queue name
		"",     // consumer name
		true,   // auto acknowledge?
		false,  // exclusive?
		false,  // no-local?
		false,  // no-wait?
		nil,    // arguments
	)
	if err != nil {
		return err
	}

	// This channel is used to block execution indefinitely
	done := make(chan struct{})

	go func() {
		defer close(done)
		for message := range messageChannel {
			var payload Payload
			err := json.Unmarshal(message.Body, &payload)
			if err != nil {
				// Handle unmarshalling error
				continue // Skip this message and continue
			}

			go handlePayload(payload)
		}
	}()

	fmt.Printf("Listening for messages [Exhange, Queue] [%s, %s]\n", EXCHANGE_NAME, q.Name)

	// Wait indefinitely
	<-done
	return nil
}

func handlePayload(payload Payload) {
	switch payload.Name {
	case "log", "event":
		err := logEvent(payload)
		if err != nil {
			log.Println(err)
		}
	// TODO: Add all possible cases
	default:
		err := logEvent(payload)
		if err != nil {
			log.Println(err)
		}
	}
}
