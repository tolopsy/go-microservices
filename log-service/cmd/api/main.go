package main

import (
	"context"
	"fmt"
	"log"
	"log-service/data"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	webPort  = "80"
	rpcPort  = "5001"
	mongoURL = "mongodb://mongo:27017"
	grpcPort = "50001"
)

var client *mongo.Client

type Config struct {
	Models data.Models
}

func main() {
	// connect to mongo
	mongo_username, mongo_password := os.Getenv("MONGO_USERNAME"), os.Getenv("MONGO_PASSWORD")
	mongoClient, err := connectToMongo(mongo_username, mongo_password)
	if err != nil {
		log.Panic(err)
	}
	client = mongoClient

	// create context to close connection
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	defer func() {
		if err := client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()

	app := Config{
		Models: data.New(client),
	}

	err = rpc.Register(new(RPCServer))
	if err != nil {
		log.Println("Error while registering RPC Server:", err.Error())
	} else {
		go func(){
			err := app.rpcListen()
			if err != nil {
				log.Println("RPC Listen Error:", err.Error())
			}
		}()
	}
	
	log.Println("Starting log service at port", webPort)
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", webPort),
		Handler: app.routes(),
	}

	err = srv.ListenAndServe()
	if err != nil {
		log.Panic(err)
	}
}

func (app *Config) rpcListen() error {
	log.Println("Starting RPC server on port ", rpcPort)
	listen, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%s", rpcPort))
	if err != nil {
		return err
	}
	defer listen.Close()
	for {
		rpcConn, err := listen.Accept()
		if err != nil {
			continue
		}

		go rpc.ServeConn(rpcConn)
	}
}

func connectToMongo(username, password string) (*mongo.Client, error) {
	clientOptions := options.Client().ApplyURI(mongoURL)
	clientOptions.SetAuth(options.Credential{
		Username: username,
		Password: password,
	})

	c, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	log.Println("Connected to Mongo!")
	return c, nil
}
