package api

import (
	"encoding/json"
	"entrytest/cmd/server/context"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"
)

type MessagesService struct {
	messageStore *context.MessageStore
}

type MessageDataRequest struct {
	Message string `json:"message"`
}

type MessageDataResponse struct {
	Id        uint64    `json:"id"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
}

func (messages *MessagesService) Post(responseWriter http.ResponseWriter, request *http.Request) {
	contentType := request.Header.Get(HeaderNameContentType)
	if contentType != HeaderValueApplicationJson {
		responseWriter.WriteHeader(http.StatusUnsupportedMediaType)
		log.Printf("[MessagesService.Post] error: unsupported media type: %s", contentType)
		return
	}

	requestBody, error := io.ReadAll(request.Body)
	if error != nil {
		responseWriter.WriteHeader(http.StatusBadRequest)
		log.Printf("[MessagesService.Post] error: failed to read request body: %v", error)
		return
	}

	var messageData MessageDataRequest
	error = json.Unmarshal(requestBody, &messageData)
	if error != nil {
		responseWriter.WriteHeader(http.StatusBadRequest)
		log.Printf("[MessagesService.Post] error: unexpected format json: %v", error)
		return
	}

	if len(messageData.Message) == 0 {
		responseWriter.WriteHeader(http.StatusBadRequest)
		log.Printf("[MessagesService.Post] error: empty message")
		return
	}

	insertedMessage := messages.messageStore.Insert(messageData.Message)

	responseBody, error := json.Marshal(MessageDataResponse(insertedMessage))
	if error != nil {
		responseWriter.WriteHeader(http.StatusInternalServerError)
		log.Printf("[MessagesService.Post] error: failed to serialize to json: %v", error)
		return
	}

	responseWriter.Header().Add(HeaderNameContentType, HeaderValueApplicationJson)
	responseWriter.Header().Add(HeaderNameContentLength, strconv.Itoa(len(responseBody)))
	responseWriter.WriteHeader(http.StatusCreated)
	written, error := responseWriter.Write(responseBody)
	if error != nil {
		log.Printf("[MessagesService.Get] error: failed to send response: %v", error)
		return
	}

	log.Printf("[MessagesService.Post] response sent, written: %d", written)
}

func (messages *MessagesService) Get(responseWriter http.ResponseWriter, request *http.Request) {
	messageList := messages.messageStore.List()
	messagePtrList := make([]*MessageDataResponse, len(messageList))
	for index := range messageList {
		messagePtrList[index] = (*MessageDataResponse)(&messageList[index])
	}

	responseBody, error := json.Marshal(messagePtrList)
	if error != nil {
		responseWriter.WriteHeader(http.StatusInternalServerError)
		log.Printf("[MessagesService.Get] error: failed to serialize to json: %v", error)
		return
	}

	responseWriter.Header().Add(HeaderNameContentType, HeaderValueApplicationJson)
	responseWriter.Header().Add(HeaderNameContentLength, strconv.Itoa(len(responseBody)))
	written, error := responseWriter.Write(responseBody)
	if error != nil {
		log.Printf("[MessagesService.Get] error: failed to send response: %v", error)
		return
	}

	log.Printf("[MessagesService.Get] response sent, written: %d", written)
}
