package api

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"
)

type EchoService struct{}

type EchoJsonData struct {
	Message string `json:"message"`
}

func (echo *EchoService) Post(responseWriter http.ResponseWriter, request *http.Request) {
	switch request.Header.Get("Content-Type") {
	case HeaderValueTextPlain:
		EchoTextPlain(responseWriter, request)
	case HeaderValueApplicationJson:
		EchoApplicationJson(responseWriter, request)
	default:
		responseWriter.WriteHeader(http.StatusUnsupportedMediaType)
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
		log.Printf("[EchoService.Post] error: failed to read request body: %v", error)
		responseWriter.WriteHeader(http.StatusBadRequest)
		return
	}

	var echoJsonData EchoJsonData
	error = json.Unmarshal(requestBody, &echoJsonData)
	if error != nil {
		log.Printf("[EchoService.Post] error: unexpected format json: %v", error)
		responseWriter.WriteHeader(http.StatusBadRequest)
		return
	}

	contentLenght := len(requestBody)
	responseWriter.Header().Add(HeaderNameContentType, HeaderValueApplicationJson)
	responseWriter.Header().Add(HeaderNameContentLength, strconv.Itoa(contentLenght))
	responseWriter.Write(requestBody)

	log.Printf("[EchoService.Post] response sent, written: %d", contentLenght)
}
