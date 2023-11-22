package main

import (
	"context"
	"fmt"
	"log"
	"log-service/data"
	"log-service/proto"
	"net"

	"google.golang.org/grpc"
)

type LogServer struct {
	proto.UnimplementedLogServiceServer
}

func (l *LogServer) WriteLog(ctx context.Context, req *proto.LogRequest) (*proto.LogResponse, error) {
	input := req.GetLogEntry()
	logEntry := data.LogEntry{
		Name: input.Name,
		Data: input.Data,
	}

	err := logEntry.Insert()
	if err != nil {
		return &proto.LogResponse{Result: "failed"}, err
	}
	result := "logged via gRPC: " + logEntry.Name

	return &proto.LogResponse{Result: result}, nil
}

func (app *Config) grpcListen() {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", grpcPort))
	if err != nil {
		log.Fatalf("Failed to listen for gRPC: %v", err)
	}

	server := grpc.NewServer()
	proto.RegisterLogServiceServer(server, &LogServer{})

	log.Printf("gRPC server started on port %s", grpcPort)
	if err := server.Serve(listener); err != nil {
		log.Fatalf("Failed to start gRPC server: %v", err)
	}
}
