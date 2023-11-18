package main

import (
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

const webPort = "80"

type Config struct {
	amqpConn *amqp.Connection
}

func main() {
	// connect to rabbitmq
	amqpConn, err := connect()
	if err != nil {
		log.Fatalln(err)
	}
	defer amqpConn.Close()
	app := Config{
		amqpConn: amqpConn,
	}

	log.Printf("Starting broker service on port %s\n", webPort)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", webPort),
		Handler: app.routes(),
	}

	err = server.ListenAndServe()
	if err != nil {
		log.Panic(err)
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
