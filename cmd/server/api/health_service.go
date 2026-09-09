package api

import (
	"log"
	"net/http"
)

type HealthService struct{}

func (health *HealthService) Get(responseWriter http.ResponseWriter, request *http.Request) {
	responseBody := []byte("ok")
	written, error := responseWriter.Write(responseBody)
	if error != nil {
		log.Printf("error: failed to send response HealthService.Get: %v, written: %d", error, written)
	}

	log.Printf("response HealthService.Get sent, written: %d", written)
}
