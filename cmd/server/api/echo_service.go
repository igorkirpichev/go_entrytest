package api

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"
)

type EchoService struct{}

type EchoDataRequest struct {
	Message string `json:"message"`
}

func (echo *EchoService) Post(responseWriter http.ResponseWriter, request *http.Request) {
	contentType := request.Header.Get(HeaderNameContentType)
	switch contentType {
	case HeaderValueTextPlain:
		EchoTextPlain(responseWriter, request)
	case HeaderValueApplicationJson:
		EchoApplicationJson(responseWriter, request)
	default:
		responseWriter.WriteHeader(http.StatusUnsupportedMediaType)
		log.Printf("[EchoService.Post] error: unsupported media type: %s", contentType)
	}

}

func EchoTextPlain(responseWriter http.ResponseWriter, request *http.Request) {
	responseWriter.Header().Add(HeaderNameContentType, HeaderValueTextPlain)
	written, error := io.Copy(responseWriter, request.Body)
	if error != nil {
		log.Printf("[EchoService.Post] error: failed to send response: %v", error)
		return
	}

	log.Printf("[EchoService.Post] response sent, written: %d", written)

}

func EchoApplicationJson(responseWriter http.ResponseWriter, request *http.Request) {
	requestBody, error := io.ReadAll(request.Body)
	if error != nil {
		responseWriter.WriteHeader(http.StatusBadRequest)
		log.Printf("[EchoService.Post] error: failed to read request body: %v", error)
		return
	}

	var echoData EchoDataRequest
	error = json.Unmarshal(requestBody, &echoData)
	if error != nil {
		responseWriter.WriteHeader(http.StatusBadRequest)
		log.Printf("[EchoService.Post] error: unexpected format json: %v", error)
		return
	}

	responseWriter.Header().Add(HeaderNameContentType, HeaderValueApplicationJson)
	responseWriter.Header().Add(HeaderNameContentLength, strconv.Itoa(len(requestBody)))
	written, error := responseWriter.Write(requestBody)
	if error != nil {
		log.Printf("[EchoService.Post] error: failed to send response: %v", error)
		return
	}

	log.Printf("[EchoService.Post] response sent, written: %d", written)
}
