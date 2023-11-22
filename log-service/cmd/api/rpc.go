package main

import (
	"context"
	"fmt"
	"log"
	"log-service/data"
	"net"
	"net/rpc"
	"time"
)

type RPCServer struct{}

type RPCPayload struct {
	Name string
	Data string
}

func (r *RPCServer) LogInfo(payload RPCPayload, status *string) error {
	collection := client.Database("logs").Collection("logs")
	_, err := collection.InsertOne(context.TODO(), data.LogEntry{
		Name:      payload.Name,
		Data:      payload.Data,
		CreatedAt: time.Now(),
	})

	if err != nil {
		log.Println("Error inserting into logs mongo collection", err)
		return err
	}

	*status = "Logged via RPC: " + payload.Name
	return nil
}

func (app *Config) rpcListen() {
	if err := rpc.Register(new(RPCServer)); err != nil {
		log.Fatalf("Error while registering RPC server: %v", err)
	}

	log.Println("Starting RPC server on port ", rpcPort)

	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", rpcPort))
	if err != nil {
		log.Fatalf("Error while listening to RPC: %v", err)
	}
	defer listener.Close()

	rpc.Accept(listener)
}
