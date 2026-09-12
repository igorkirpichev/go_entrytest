package main

import (
	"entrytest/cmd/server/api"
	"entrytest/cmd/server/context"
	"log"
	"net/http"
	"os"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	messageStore := context.CreateMessageStore()

	serverApi := api.CreateServerApi(messageStore)
	mux := http.NewServeMux()

	mux.Handle("/", http.FileServer(http.Dir("frontend")))
	mux.HandleFunc("GET /health", serverApi.Health.Get)
	mux.HandleFunc("POST /echo", serverApi.Echo.Post)
	mux.HandleFunc("POST /messages", serverApi.Messages.Post)
	mux.HandleFunc("GET /messages", serverApi.Messages.Get)
	mux.HandleFunc("DELETE /messages/{id}", serverApi.Messages.Delete)

	log.Printf("сервер слушает http://localhost:%s", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}
