package service

type HealthService struct{}

func NewHealthService() *HealthService {
	return &HealthService{}
}

func (s *HealthService) GetPingMessage() string {
	return "Medislot API is running"
}
