# GO Microservices

This repository demonstrates various communication patterns and protocols for microservices in Go.

## Overview

The repository includes five microservices:

- **Broker service:** Handles request distribution among other microservices. It utilizes an AMQP emitter to dispatch events to RabbitMQ for asynchronous handling. Capable of forwarding log requests to the log service via REST, RPC, and gRPC or to RabbitMQ via AMQP.
- **Authentication service**
- **Mail service**
- **Listener service:** Monitors AMQP events in the RabbitMQ default queue.
- **Log service:** Manages logs, storing them in a MongoDB collection. Accepts log requests through REST, RPC, or gRPC.

The repository also includes a frontend app with a user interface for testing all microservices.

## Pre-requisites

- Docker
- Make utility

## Quick Start Guide

1. Navigate to the 'infra' directory: `cd infra`
2. Build and start all infrastructure services: `make up_build`
3. Launch the frontend app: `make start`
4. Access the frontend UI at http://localhost

Note: The connection to the broker-amqp might take a few seconds to establish, as RabbitMQ may need time to initialize.

## Run Commands

(Ensure the current working directory is 'infra'.)

- `make up_build`: Build/re-build and start/restart containers for all services in the infra docker-compose file.
- `make up`: Start containers for all infra services without rebuilding.
- `make down`: Remove all infra containers.
- `make start`: Start the frontend app.
- `make stop`: Stop the frontend app.

All commands are implemented in the infra's Makefile.
