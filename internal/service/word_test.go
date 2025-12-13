package service

import (
	"fmt"
	"testing"
	"time"

	"languager/internal/domain"
	"languager/internal/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestWordService_SaveWordPair(t *testing.T) {
	tests := []struct {
		name          string
		userID        int64
		word          string
		translation   string
		mockError     error
		expectedError bool
	}{
		{
			name:          "valid word pair",
			userID:        123,
			word:          "hello",
			translation:   "привет",
			mockError:     nil,
			expectedError: false,
		},
		{
			name:          "empty word",
			userID:        123,
			word:          "",
			translation:   "привет",
			mockError:     nil,
			expectedError: true,
		},
		{
			name:          "empty translation",
			userID:        123,
			word:          "hello",
			translation:   "",
			mockError:     nil,
			expectedError: true,
		},
		{
			name:          "both empty",
			userID:        123,
			word:          "",
			translation:   "",
			mockError:     nil,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(testutil.MockWordRepository)

			// Only set up mock if inputs are valid
			if tt.word != "" && tt.translation != "" {
				mockRepo.On("SaveWord", tt.userID, tt.word, tt.translation).Return(tt.mockError)
			}

			service := NewWordService(mockRepo)

			err := service.SaveWordPair(tt.userID, tt.word, tt.translation)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if tt.word != "" && tt.translation != "" {
				mockRepo.AssertExpectations(t)
			}
		})
	}
}

func TestWordService_GetRandomPair(t *testing.T) {
	testWord := testutil.NewTestWord(1, 123, "hello", "привет")

	tests := []struct {
		name          string
		userID        int64
		mockReturn    *domain.Word
		mockError     error
		expectedWord  *domain.Word
		expectedError bool
	}{
		{
			name:          "word found",
			userID:        123,
			mockReturn:    testWord,
			mockError:     nil,
			expectedWord:  testWord,
			expectedError: false,
		},
		{
			name:          "no words",
			userID:        456,
			mockReturn:    nil,
			mockError:     nil,
			expectedWord:  nil,
			expectedError: false,
		},
		{
			name:          "database error",
			userID:        789,
			mockReturn:    nil,
			mockError:     fmt.Errorf("db error"),
			expectedWord:  nil,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(testutil.MockWordRepository)
			mockRepo.On("GetRandomWord", tt.userID).Return(tt.mockReturn, tt.mockError)

			service := NewWordService(mockRepo)

			word, err := service.GetRandomPair(tt.userID)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedWord, word)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestWordService_GetDaysList(t *testing.T) {
	tests := []struct {
		name              string
		userID            int64
		page              int
		mockDays          []domain.Day
		mockTotalDays     int
		mockError         error
		mockTotalDaysError error
		expectedPages     int
		expectedDaysCount int
		expectedError     bool
	}{
		{
			name:              "first page with 7 days",
			userID:            123,
			page:              1,
			mockDays:          []domain.Day{testutil.NewTestDay(time.Now(), 5), testutil.NewTestDay(time.Now().AddDate(0, 0, -1), 3)},
			mockTotalDays:     14,
			mockError:         nil,
			mockTotalDaysError: nil,
			expectedPages:     2,
			expectedDaysCount: 2,
			expectedError:     false,
		},
		{
			name:              "invalid page number (negative)",
			userID:            123,
			page:              -1,
			mockDays:          []domain.Day{},
			mockTotalDays:     7,
			mockError:         nil,
			mockTotalDaysError: nil,
			expectedPages:     1,
			expectedDaysCount: 0,
			expectedError:     false,
		},
		{
			name:              "page zero defaults to 1",
			userID:            123,
			page:              0,
			mockDays:          []domain.Day{testutil.NewTestDay(time.Now(), 5)},
			mockTotalDays:     1,
			mockError:         nil,
			mockTotalDaysError: nil,
			expectedPages:     1,
			expectedDaysCount: 1,
			expectedError:     false,
		},
		{
			name:              "database error on days",
			userID:            123,
			page:              1,
			mockError:         fmt.Errorf("db error"),
			mockTotalDaysError: nil,
			expectedError:     true,
		},
		{
			name:              "zero total days sets totalPages to 1",
			userID:            123,
			page:              1,
			mockDays:          []domain.Day{},
			mockTotalDays:     0,
			mockError:         nil,
			mockTotalDaysError: nil,
			expectedPages:     1,
			expectedDaysCount: 0,
			expectedError:     false,
		},
		{
			name:              "database error on total count",
			userID:            123,
			page:              1,
			mockDays:          []domain.Day{testutil.NewTestDay(time.Now(), 5)},
			mockError:          nil,
			mockTotalDaysError: fmt.Errorf("db error"),
			expectedError:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(testutil.MockWordRepository)

			page := tt.page
			if page < 1 {
				page = 1
			}
			offset := (page - 1) * 7

			mockRepo.On("GetDaysWithWords", tt.userID, 7, offset).Return(tt.mockDays, tt.mockError)

			if tt.mockError == nil {
				if tt.mockTotalDaysError != nil {
					mockRepo.On("GetTotalDaysCount", tt.userID).Return(0, tt.mockTotalDaysError)
				} else {
					mockRepo.On("GetTotalDaysCount", tt.userID).Return(tt.mockTotalDays, nil)
				}
			}

			service := NewWordService(mockRepo)

			days, totalPages, err := service.GetDaysList(tt.userID, tt.page)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedPages, totalPages)
				assert.Len(t, days, tt.expectedDaysCount)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestWordService_GetWordsByDate(t *testing.T) {
	tests := []struct {
		name          string
		dateStr       string
		mockWords     []domain.Word
		mockError     error
		expectedError bool
	}{
		{
			name:    "valid date",
			dateStr: "20241212",
			mockWords: []domain.Word{
				*testutil.NewTestWord(1, 123, "hello", "привет"),
			},
			mockError:     nil,
			expectedError: false,
		},
		{
			name:          "invalid date format",
			dateStr:       "2024-12-12",
			mockError:     nil,
			expectedError: true,
		},
		{
			name:          "empty date",
			dateStr:       "",
			mockError:     nil,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(testutil.MockWordRepository)

			if !tt.expectedError {
				date, _ := time.Parse("20060102", tt.dateStr)
				mockRepo.On("GetWordsByDate", int64(123), mock.MatchedBy(func(d time.Time) bool {
					return d.Year() == date.Year() && d.Month() == date.Month() && d.Day() == date.Day()
				})).Return(tt.mockWords, tt.mockError)
			}

			service := NewWordService(mockRepo)

			words, err := service.GetWordsByDate(123, tt.dateStr)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.mockWords, words)
			}

			if !tt.expectedError {
				mockRepo.AssertExpectations(t)
			}
		})
	}
}

func TestWordService_HideWordFor7Days(t *testing.T) {
	tests := []struct {
		name          string
		wordID        int
		mockError     error
		expectedError bool
	}{
		{
			name:          "successful hide",
			wordID:        1,
			mockError:     nil,
			expectedError: false,
		},
		{
			name:          "database error",
			wordID:        2,
			mockError:     fmt.Errorf("database error"),
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(testutil.MockWordRepository)
			mockRepo.On("HideWordFor7Days", tt.wordID).Return(tt.mockError)

			service := NewWordService(mockRepo)

			err := service.HideWordFor7Days(tt.wordID)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestWordService_HideWordForever(t *testing.T) {
	tests := []struct {
		name          string
		wordID        int
		mockError     error
		expectedError bool
	}{
		{
			name:          "successful hide",
			wordID:        1,
			mockError:     nil,
			expectedError: false,
		},
		{
			name:          "database error",
			wordID:        2,
			mockError:     fmt.Errorf("database error"),
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(testutil.MockWordRepository)
			mockRepo.On("HideWordForever", tt.wordID).Return(tt.mockError)

			service := NewWordService(mockRepo)

			err := service.HideWordForever(tt.wordID)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

