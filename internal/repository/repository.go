package repository

import (
	"time"

	"languager/internal/domain"
)

// UserRepository defines user data operations
type UserRepository interface {
	IsAuthorized(userID int64) (bool, error)
	AuthorizeUser(userID int64) error
	EnsureUserExists(userID int64) error
}

// WordRepository defines word data operations
type WordRepository interface {
	SaveWord(userID int64, word, translation string) error
	GetRandomWord(userID int64) (*domain.Word, error)
	GetDaysWithWords(userID int64, limit, offset int) ([]domain.Day, error)
	GetWordsByDate(userID int64, date time.Time) ([]domain.Word, error)
	CleanOldWords(days int) error
	GetTotalDaysCount(userID int64) (int, error)
}

