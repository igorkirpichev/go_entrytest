package api

type ServerApi struct {
	Health HealthService
	Echo   EchoService
}

func CreateServerApi() *ServerApi {
	serverApi := ServerApi{}
	return &serverApi
}
