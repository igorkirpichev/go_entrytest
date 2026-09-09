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
		log.Printf("[HealthService.Get] error: failed to send response: %v", error)
		return
	}

	log.Printf("[HealthService.Get] response sent, written: %d", written)
}
