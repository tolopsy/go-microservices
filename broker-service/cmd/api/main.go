package main

import (
	"fmt"
	"log"
	"net/http"
)

const webPort = "80"

type Config struct {}

func main() {
	// TODO: To be used or removed
	app := Config{}

	log.Printf("Starting broker service on port %s\n", webPort)

	server := &http.Server{
		Addr: fmt.Sprintf(":%s", webPort),
		Handler: app.routes(),
	}

	err := server.ListenAndServe()
	if err != nil {
		log.Panic(err)
	}
}