package main

import (
	"broker/proto"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/rpc"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type RequestPayload struct {
	Action string      `json:"action"`
	Auth   AuthPayload `json:"auth,omitempty"`
	Log    LogPayload  `json:"log,omitempty"`
	Mail   MailPayload `json:"mail,omitempty"`
}

type AuthPayload struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LogPayload struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

type MailPayload struct {
	From    string `json:"from"`
	To      string `json:"to"`
	Subject string `json:"subject"`
	Message string `json:"message"`
}

type RPCPayload struct {
	Name string
	Data string
}

func (app *Config) Broker(w http.ResponseWriter, r *http.Request) {
	payload := jsonResponse{
		Error:   false,
		Message: "Hit the broker",
	}
	app.writeJson(w, http.StatusOK, payload)
}

func (app *Config) HandleRequest(w http.ResponseWriter, r *http.Request) {
	var requestPayload RequestPayload

	err := app.readJson(w, r, &requestPayload)
	if err != nil {
		app.errorJson(w, err)
		return
	}

	switch requestPayload.Action {
	case "auth":
		app.authenticate(w, requestPayload.Auth)
	case "log":
		app.logItem(w, requestPayload.Log)

	case "logAMQP":
		app.logItemViaAMQP(w, requestPayload.Log)

	case "logRPC":
		app.logItemViaRPC(w, requestPayload.Log)

	case "logGRPC":
		app.logItemViaGRPC(w, requestPayload.Log)

	case "mail":
		app.sendMail(w, requestPayload.Mail)

	default:
		app.errorJson(w, errors.New("unknown action"))
	}
}

func (app *Config) logItem(w http.ResponseWriter, entry LogPayload) {
	jsonData, _ := json.Marshal(entry)
	request, err := http.NewRequest("POST", "http://log-service/log", bytes.NewBuffer(jsonData))
	if err != nil {
		app.errorJson(w, err)
		return
	}

	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	response, err := client.Do(request)

	if err != nil {
		app.errorJson(w, err)
		return
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusAccepted {
		app.errorJson(w, errors.New("error calling auth service"))
		return
	}

	payload := jsonResponse{
		Error:   false,
		Message: "logged",
	}

	app.writeJson(w, http.StatusAccepted, payload)
}

func (app *Config) authenticate(w http.ResponseWriter, a AuthPayload) {
	jsonData, _ := json.Marshal(a)
	request, err := http.NewRequest("POST", "http://authentication-service/authenticate", bytes.NewBuffer(jsonData))
	if err != nil {
		app.errorJson(w, err)
		return
	}

	client := &http.Client{}
	response, err := client.Do(request)

	if err != nil {
		app.errorJson(w, err)
		return
	}
	defer response.Body.Close()

	if response.StatusCode == http.StatusUnauthorized {
		app.errorJson(w, errors.New("invalid credentials"))
		return
	} else if response.StatusCode != http.StatusAccepted {
		app.errorJson(w, errors.New("error calling auth service"))
		return
	}

	var jsonFromAuthService jsonResponse
	err = json.NewDecoder(response.Body).Decode(&jsonFromAuthService)
	if err != nil {
		app.errorJson(w, err)
		return
	}

	if jsonFromAuthService.Error {
		app.errorJson(w, err, http.StatusUnauthorized)
		return
	}

	payload := jsonResponse{
		Error:   false,
		Message: "Authenticated",
		Data:    jsonFromAuthService.Data,
	}

	app.writeJson(w, http.StatusAccepted, payload)
}

func (app *Config) sendMail(w http.ResponseWriter, m MailPayload) {
	jsonData, _ := json.Marshal(m)
	request, err := http.NewRequest("POST", "http://mail-service/send", bytes.NewBuffer(jsonData))
	if err != nil {
		app.errorJson(w, err)
		return
	}

	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	response, err := client.Do(request)

	if err != nil {
		app.errorJson(w, err)
		return
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusAccepted {
		app.errorJson(w, errors.New("error calling mail service"))
		return
	}

	payload := jsonResponse{
		Error:   false,
		Message: "message sent to " + m.To,
	}

	app.writeJson(w, http.StatusAccepted, payload)
}

func (app *Config) logItemViaAMQP(w http.ResponseWriter, l LogPayload) {
	err := app.pushToQueue(l, "log.INFO")
	if err != nil {
		app.errorJson(w, err)
		return
	}

	payload := jsonResponse{
		Error:   false,
		Message: "logged via AMQP (RabbitMQ)",
	}

	app.writeJson(w, http.StatusAccepted, payload)
}

func (app *Config) logItemViaRPC(w http.ResponseWriter, l LogPayload) {
	client, err := rpc.Dial("tcp", "log-service:5001")
	if err != nil {
		app.errorJson(w, err)
		return
	}

	rpcPayload := RPCPayload(l)

	var result string
	err = client.Call("RPCServer.LogInfo", rpcPayload, &result)
	if err != nil {
		app.errorJson(w, err)
		return
	}

	responsePayload := jsonResponse{
		Error:   false,
		Message: result,
	}
	app.writeJson(w, http.StatusAccepted, responsePayload)
}

func (app *Config) logItemViaGRPC(w http.ResponseWriter, l LogPayload) {
	conn, err := grpc.Dial(
		"log-service:50001",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		app.errorJson(w, err)
	}
	defer conn.Close()

	client := proto.NewLogServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	logResponse, err := client.WriteLog(ctx, &proto.LogRequest{
		LogEntry: &proto.Log{
			Name: l.Name,
			Data: l.Data,
		},
	})
	if err != nil {
		app.errorJson(w, err)
	}

	var payload = jsonResponse{
		Error:   false,
		Message: logResponse.GetResult(),
	}
	app.writeJson(w, http.StatusAccepted, payload)
}
