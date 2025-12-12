package service

import (
	"languager/internal/repository"
)

// AuthService handles authentication logic
type AuthService struct {
	userRepo    repository.UserRepository
	botPassword string
}

// NewAuthService creates a new auth service
func NewAuthService(userRepo repository.UserRepository, botPassword string) *AuthService {
	return &AuthService{
		userRepo:    userRepo,
		botPassword: botPassword,
	}
}

// CheckPassword verifies if provided password matches
func (s *AuthService) CheckPassword(password string) bool {
	return password == s.botPassword
}

// IsAuthorized checks if user is authorized
func (s *AuthService) IsAuthorized(userID int64) (bool, error) {
	return s.userRepo.IsAuthorized(userID)
}

// AuthorizeUser authorizes a user
func (s *AuthService) AuthorizeUser(userID int64) error {
	return s.userRepo.AuthorizeUser(userID)
}

// EnsureUserExists creates user record if doesn't exist
func (s *AuthService) EnsureUserExists(userID int64) error {
	return s.userRepo.EnsureUserExists(userID)
}

