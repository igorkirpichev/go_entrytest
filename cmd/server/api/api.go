package api

import "entrytest/cmd/server/context"

type ServerApi struct {
	Health   HealthService
	Echo     EchoService
	Messages MessagesService
}

func CreateServerApi(messageStore *context.MessageStore) *ServerApi {
	serverApi := ServerApi{}
	serverApi.Messages.messageStore = messageStore

	return &serverApi
}
