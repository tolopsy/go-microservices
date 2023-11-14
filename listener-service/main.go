package main

import (
	"log"
	"math"
	"os"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	// connect to rabbitmq
	conn, err := connect()
	if err != nil {
		log.Fatalln(err)
	}
	defer conn.Close()
	log.Println("Connected to RabbitMQ")

	// start listening for messages
	// create consumer
	// watch the queue and consume events
}

// Connecting to the RabbitMQ with exponential backoff retry until successful or maximum retries are reached.
func connect() (*amqp.Connection, error) {
	AMQP_URL := os.Getenv("AMQP_URL")
	var retryCount int64
	const maxRetries = 5

	for {
		connection, err := amqp.Dial(AMQP_URL)
		if err == nil {
			return connection, nil
		}

		retryCount++
		if retryCount > maxRetries {
			return nil, err
		}

		backoff := time.Duration(math.Pow(float64(retryCount), 2)) * time.Second
		log.Printf("Backing off! Retrying in %s", backoff)
		time.Sleep(backoff)
		continue
	}
}
