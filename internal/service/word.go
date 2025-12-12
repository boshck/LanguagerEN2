package service

import (
	"fmt"
	"time"

	"languager/internal/domain"
	"languager/internal/repository"
)

// WordService handles word-related business logic
type WordService struct {
	wordRepo repository.WordRepository
}

// NewWordService creates a new word service
func NewWordService(wordRepo repository.WordRepository) *WordService {
	return &WordService{wordRepo: wordRepo}
}

// SaveWordPair saves a word-translation pair
func (s *WordService) SaveWordPair(userID int64, word, translation string) error {
	if word == "" || translation == "" {
		return fmt.Errorf("word and translation cannot be empty")
	}
	return s.wordRepo.SaveWord(userID, word, translation)
}

// GetRandomPair returns a random word-translation pair
func (s *WordService) GetRandomPair(userID int64) (*domain.Word, error) {
	return s.wordRepo.GetRandomWord(userID)
}

// GetDaysList returns paginated list of days with word counts
func (s *WordService) GetDaysList(userID int64, page int) ([]domain.Day, int, error) {
	const pageSize = 7

	if page < 1 {
		page = 1
	}

	offset := (page - 1) * pageSize
	days, err := s.wordRepo.GetDaysWithWords(userID, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}

	// Calculate total pages
	totalDays, err := s.wordRepo.GetTotalDaysCount(userID)
	if err != nil {
		return nil, 0, err
	}

	totalPages := (totalDays + pageSize - 1) / pageSize
	if totalPages == 0 {
		totalPages = 1
	}

	return days, totalPages, nil
}

// GetWordsByDate returns all words for a specific date
func (s *WordService) GetWordsByDate(userID int64, dateStr string) ([]domain.Word, error) {
	// Parse date string (YYYYMMDD format)
	date, err := time.Parse("20060102", dateStr)
	if err != nil {
		return nil, fmt.Errorf("invalid date format: %w", err)
	}

	return s.wordRepo.GetWordsByDate(userID, date)
}

