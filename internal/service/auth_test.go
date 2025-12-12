package service

import (
	"testing"

	"languager/internal/testutil"

	"github.com/stretchr/testify/assert"
)

func TestAuthService_CheckPassword(t *testing.T) {
	tests := []struct {
		name           string
		botPassword    string
		inputPassword  string
		expectedResult bool
	}{
		{
			name:           "correct password",
			botPassword:    "secret123",
			inputPassword:  "secret123",
			expectedResult: true,
		},
		{
			name:           "incorrect password",
			botPassword:    "secret123",
			inputPassword:  "wrong",
			expectedResult: false,
		},
		{
			name:           "empty password",
			botPassword:    "secret123",
			inputPassword:  "",
			expectedResult: false,
		},
		{
			name:           "case sensitive",
			botPassword:    "Secret123",
			inputPassword:  "secret123",
			expectedResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(testutil.MockUserRepository)
			service := NewAuthService(mockRepo, tt.botPassword)

			result := service.CheckPassword(tt.inputPassword)

			assert.Equal(t, tt.expectedResult, result)
		})
	}
}

func TestAuthService_IsAuthorized(t *testing.T) {
	tests := []struct {
		name          string
		userID        int64
		mockReturn    bool
		mockError     error
		expectedAuth  bool
		expectedError bool
	}{
		{
			name:          "authorized user",
			userID:        123,
			mockReturn:    true,
			mockError:     nil,
			expectedAuth:  true,
			expectedError: false,
		},
		{
			name:          "unauthorized user",
			userID:        456,
			mockReturn:    false,
			mockError:     nil,
			expectedAuth:  false,
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(testutil.MockUserRepository)
			mockRepo.On("IsAuthorized", tt.userID).Return(tt.mockReturn, tt.mockError)

			service := NewAuthService(mockRepo, "password")

			authorized, err := service.IsAuthorized(tt.userID)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedAuth, authorized)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestAuthService_AuthorizeUser(t *testing.T) {
	mockRepo := new(testutil.MockUserRepository)
	mockRepo.On("AuthorizeUser", int64(123)).Return(nil)

	service := NewAuthService(mockRepo, "password")

	err := service.AuthorizeUser(123)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestAuthService_EnsureUserExists(t *testing.T) {
	mockRepo := new(testutil.MockUserRepository)
	mockRepo.On("EnsureUserExists", int64(123)).Return(nil)

	service := NewAuthService(mockRepo, "password")

	err := service.EnsureUserExists(123)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

