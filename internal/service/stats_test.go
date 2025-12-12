package service

import (
	"fmt"
	"testing"

	"languager/internal/testutil"

	"github.com/stretchr/testify/assert"
)

func TestStatsService_CleanupOldData(t *testing.T) {
	tests := []struct {
		name          string
		mockError     error
		expectedError bool
	}{
		{
			name:          "successful cleanup",
			mockError:     nil,
			expectedError: false,
		},
		{
			name:          "database error",
			mockError:     fmt.Errorf("db error"),
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(testutil.MockWordRepository)
			mockRepo.On("CleanOldWords", 60).Return(tt.mockError)

			logger := testutil.NewTestLogger()
			service := NewStatsService(mockRepo, logger)

			err := service.CleanupOldData()

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

