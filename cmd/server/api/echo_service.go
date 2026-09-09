package api

import (
	"io"
	"log"
	"net/http"
)

type EchoService struct{}

func (echo *EchoService) Post(responseWriter http.ResponseWriter, request *http.Request) {
	switch request.Header.Get("Content-Type") {
	case "text/plain":
		EchoTextPlain(responseWriter, request)
	default:
		responseWriter.WriteHeader(http.StatusUnsupportedMediaType)
	}

}

func EchoTextPlain(responseWriter http.ResponseWriter, request *http.Request) {
	written, error := io.Copy(responseWriter, request.Body)
	if error != nil {
		log.Printf("error: failed to send response EchoService.Post: %v, written: %d", error, written)
	}

	log.Printf("response EchoService.Post sent, written: %d", written)

}
