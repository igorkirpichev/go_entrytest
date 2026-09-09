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
	Id        uint32    `json:"id"`
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

	contentLenght := len(responseBody)
	responseWriter.Header().Add(HeaderNameContentType, HeaderValueApplicationJson)
	responseWriter.Header().Add(HeaderNameContentLength, strconv.Itoa(contentLenght))
	responseWriter.WriteHeader(http.StatusCreated)
	responseWriter.Write(responseBody)

	log.Printf("[MessagesService.Post] response sent, written: %d", contentLenght)
}
