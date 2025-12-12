package testutil

import (
	"languager/internal/domain"
	"time"

	"go.uber.org/zap"
)

// NewTestLogger creates a no-op logger for tests
func NewTestLogger() *zap.Logger {
	return zap.NewNop()
}

// NewTestUser creates a test user
func NewTestUser(userID int64, authorized bool) *domain.User {
	return &domain.User{
		UserID:     userID,
		Authorized: authorized,
		CreatedAt:  time.Now(),
	}
}

// NewTestWord creates a test word
func NewTestWord(id int, userID int64, word, translation string) *domain.Word {
	return &domain.Word{
		ID:          id,
		UserID:      userID,
		Word:        word,
		Translation: translation,
		CreatedAt:   time.Now(),
	}
}

// NewTestDay creates a test day
func NewTestDay(date time.Time, wordCount int) domain.Day {
	return domain.Day{
		Date:      date,
		WordCount: wordCount,
	}
}

