package main

import (
	"log"
	"net/http"

	"github.com/tls1641/ws/internal/handlers"
)

func main() {

	mux := routes()

	log.Println("starting channel listener")

	go handlers.ListenToWsChannel()

	log.Println("Starting web server on port 8080")

	_ = http.ListenAndServe(":8080", mux)
}
