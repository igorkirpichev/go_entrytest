package api

type ServerApi struct {
	Health HealthService
}

func CreateServerApi() *ServerApi {
	serverApi := ServerApi{}
	return &serverApi
}
