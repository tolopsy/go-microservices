package main

import (
	"authentication/data"
	"database/sql"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"time"

	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"
)

const webPort = "80"

type Config struct {
	DB     *sql.DB
	Models data.Models
}

func main() {
	log.Println("Starting authentication service...")

	// Connect to DB
	conn := connectToDB()
	if conn == nil {
		log.Panic("Can't connect to Postgres!")
	}

	// set up config
	app := Config{
		DB:     conn,
		Models: data.New(conn),
	}

	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", webPort),
		Handler: app.routes(),
	}

	err := server.ListenAndServe()
	if err != nil {
		log.Panic(err)
	}
}

func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

func connectToDB() *sql.DB {
	dsn := os.Getenv("DSN")
	retryCount := 0
	maxRetries := 10

	// Retry connecting to the database until successful or maximum retries are reached.
	for {
		connection, err := openDB(dsn)
		if err == nil {
			log.Println("Connected to Postgres!")
			return connection
		}

		retryCount++
		if retryCount > maxRetries {
			log.Println(err)
			return nil
		}

		backoff := time.Duration(math.Pow(float64(retryCount), 2)) * time.Second
		log.Printf("Backing off! Retrying in %s", backoff)
		time.Sleep(backoff)
		continue
	}
}
