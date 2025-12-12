package service

import (
	"languager/internal/repository"

	"go.uber.org/zap"
)

// StatsService handles statistics and cleanup
type StatsService struct {
	wordRepo repository.WordRepository
	logger   *zap.Logger
}

// NewStatsService creates a new stats service
func NewStatsService(wordRepo repository.WordRepository, logger *zap.Logger) *StatsService {
	return &StatsService{
		wordRepo: wordRepo,
		logger:   logger,
	}
}

// CleanupOldData removes words older than 60 days
func (s *StatsService) CleanupOldData() error {
	const retentionDays = 60

	s.logger.Info("Starting cleanup of old words", zap.Int("retention_days", retentionDays))

	err := s.wordRepo.CleanOldWords(retentionDays)
	if err != nil {
		s.logger.Error("Failed to cleanup old words", zap.Error(err))
		return err
	}

	s.logger.Info("Cleanup completed successfully")
	return nil
}

