package services

import (
	"github.com/Code-Aether/americanas-loja-api/internal/repository"
)

type AuthService struct {
	userRepo *repository.UserRepository
}

func NewAuthService(userRepo *repository.UserRepository) *AuthService {
	return &AuthService{
		userRepo: userRepo,
	}
}

func (s *AuthService) Login(email, password string) (string, error) {
	return "fake-jwt-token", nil // TODO: Implement JWT token generation
}

func (s *AuthService) Register(email, password string) error {
	return nil
}
