package testutil

import (
	"languager/internal/domain"
	"time"

	"github.com/stretchr/testify/mock"
)

// MockUserRepository is a mock for UserRepository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) IsAuthorized(userID int64) (bool, error) {
	args := m.Called(userID)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserRepository) AuthorizeUser(userID int64) error {
	args := m.Called(userID)
	return args.Error(0)
}

func (m *MockUserRepository) EnsureUserExists(userID int64) error {
	args := m.Called(userID)
	return args.Error(0)
}

// MockWordRepository is a mock for WordRepository
type MockWordRepository struct {
	mock.Mock
}

func (m *MockWordRepository) SaveWord(userID int64, word, translation string) error {
	args := m.Called(userID, word, translation)
	return args.Error(0)
}

func (m *MockWordRepository) GetRandomWord(userID int64) (*domain.Word, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Word), args.Error(1)
}

func (m *MockWordRepository) GetDaysWithWords(userID int64, limit, offset int) ([]domain.Day, error) {
	args := m.Called(userID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Day), args.Error(1)
}

func (m *MockWordRepository) GetWordsByDate(userID int64, date time.Time) ([]domain.Word, error) {
	args := m.Called(userID, date)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Word), args.Error(1)
}

func (m *MockWordRepository) CleanOldWords(days int) error {
	args := m.Called(days)
	return args.Error(0)
}

func (m *MockWordRepository) GetTotalDaysCount(userID int64) (int, error) {
	args := m.Called(userID)
	return args.Int(0), args.Error(1)
}

