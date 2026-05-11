package service

import (
	"log/slog"

	"medislot/internal/models"
	"medislot/internal/repository"
)

type StatsService interface {
	GetSystemStats() (*models.SystemStats, error)
}

type statsService struct{ statsRepo repository.StatsRepository }

func NewStatsService(statsRepo repository.StatsRepository) StatsService {
	return &statsService{statsRepo: statsRepo}
}

func (s *statsService) GetSystemStats() (*models.SystemStats, error) {
	stats, err := s.statsRepo.GetSystemStats()
	if err != nil {
		slog.Error("get system stats failed", "error", err)
		return nil, err
	}
	return stats, nil
}
