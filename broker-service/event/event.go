package event

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

const EXCHANGE_NAME = "logs_topic"
const publishTimeout = time.Second * 5

type Payload struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

func declareExchange(channel *amqp.Channel) error {
	return channel.ExchangeDeclare(
		EXCHANGE_NAME, // name
		"topic",       // type
		true,          // durable?
		false,         // auto delete?
		false,         // internal?
		false,         // no-wait?
		nil,           // args
	)
}

func declareQueue(channel *amqp.Channel) (amqp.Queue, error) {
	return channel.QueueDeclare(
		"",    // name
		false, // durable?
		false, // auto delete?
		true,  // exclusive,
		false, // no-wait?
		nil,   // args
	)
}

func logEvent(payload Payload) error {
	jsonData, _ := json.Marshal(payload)
	request, err := http.NewRequest("POST", "http://log-service/log", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	response, err := client.Do(request)

	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusAccepted {
		return err
	}

	return nil
}
