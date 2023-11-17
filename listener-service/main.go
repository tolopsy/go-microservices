package main

import (
	"listener-service/event"
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

	// create consumer
	consumer, err := event.NewConsumer(conn)
	if err != nil {
		panic(err)
	}

	// watch the queue and consume events
	err = consumer.Listen([]string{"log.INFO", "log.WARNING", "log.ERROR"})
	if err != nil {
		log.Println(err)
	}
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
