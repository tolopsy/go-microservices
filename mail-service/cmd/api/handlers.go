package main

import "net/http"

func (app *Config) SendMail(w http.ResponseWriter, r *http.Request) {
	var requestPayload struct{
		From string `json:"from"`
		To string `json:"to"`
		Subject string `json:"subject"`
		Message string `json:"message"`
	}

	err := app.readJson(w, r, &requestPayload)
	if err != nil {
		app.errorJson(w, err)
		return
	}

	msg := Message{
		From: requestPayload.From,
		To: requestPayload.To,
		Subject: requestPayload.Subject,
		Data: requestPayload.Message,
	}

	err = app.Mailer.SendSMTPMessage(msg)
	if err != nil {
		app.errorJson(w, err)
		return
	}

	payload := jsonResponse{
		Error: false,
		Message: "sent to " + requestPayload.To,
	}

	app.writeJson(w, http.StatusAccepted, payload)
}