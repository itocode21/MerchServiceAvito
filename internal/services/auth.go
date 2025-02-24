package services

import (
	"fmt"
	"log"

	"github.com/itocode21/MerchServiceAvito/internal/auth"
	"github.com/itocode21/MerchServiceAvito/internal/repositories"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepo *repositories.UserRepository
}

func NewAuthService(userRepo *repositories.UserRepository) *AuthService {
	return &AuthService{userRepo: userRepo}
}

func (s *AuthService) Authenticate(username, password string) (string, error) {
	user, err := s.userRepo.GetUserByUsername(username)
	if err != nil {
		return "", fmt.Errorf("ошибка при получении пользователя: %v", err)
	}
	if user == nil {
		return "", fmt.Errorf("пользователь не найден")
	}

	log.Printf("Password from request: %s", password)
	log.Printf("PasswordHash from DB: %s", user.PasswordHash)
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		log.Printf("CompareHashAndPassword error: %v", err)
		return "", fmt.Errorf("неверный пароль")
	}

	token, err := auth.GenerateJWT(user.Username)
	if err != nil {
		return "", fmt.Errorf("ошибка генерации токена: %v", err)
	}
	return token, nil
}
