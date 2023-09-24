package main

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

type jsonResponse struct {
	Error   bool        `json:"error"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func (app *Config) readJson(w http.ResponseWriter, r *http.Request, data interface{}) error {
	// maximum limit cast on json data
	maxBytes := 1048576 // 1 MB

	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))
	dec := json.NewDecoder(r.Body)
	err := dec.Decode(data)
	if err != nil {
		return err
	}

	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		return errors.New("Request body should consist of a single JSON value, not an array of values.")
	}

	return nil
}

func (app *	Config) writeJson(w http.ResponseWriter, status int, data interface{}, headerMaps ...map[string]string) error {
	out, err := json.Marshal(data)
	if err != nil {
		return err
	}
	if len(headerMaps) > 0 {
		for key, value := range headerMaps[0] {
			w.Header().Add(key, value)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	_, err = w.Write(out)
	if err != nil {
		return err
	}

	return err
}

func (app *Config) errorJson(w http.ResponseWriter, r *http.Request, err error, status ...int) error {
	statusCode := http.StatusBadRequest
	if len(status) > 0 {
		statusCode = status[0]
	}

	payload := jsonResponse{
		Error: true,
		Message: err.Error(),
	}

	return app.writeJson(w, statusCode, payload)
}