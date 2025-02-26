package services

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

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
	cacheKey := "user_hash:" + username
	cachedHash, err := s.userRepo.Config.Redis.Get(context.Background(), cacheKey).Result()
	if err == nil {
		if err := bcrypt.CompareHashAndPassword([]byte(cachedHash), []byte(password)); err == nil {
			return auth.GenerateJWT(username)
		}
	}

	user, err := s.userRepo.GetUserByUsername(username)
	if err != nil {
		return "", fmt.Errorf("ошибка при получении пользователя: %v", err)
	}
	if user == nil {
		return "", fmt.Errorf("пользователь не найден")
	}

	var wg sync.WaitGroup
	var hashErr error
	log.Printf("Password from request: %s", password)
	log.Printf("PasswordHash from DB: %s", user.PasswordHash)

	wg.Add(1)
	go func() {
		defer wg.Done()
		hashErr = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	}()

	wg.Wait()
	if hashErr != nil {
		log.Printf("CompareHashAndPassword error: %v", hashErr)
		return "", fmt.Errorf("неверный пароль")
	}

	s.userRepo.Config.Redis.Set(context.Background(), cacheKey, user.PasswordHash, 5*time.Minute)

	token, err := auth.GenerateJWT(user.Username)
	if err != nil {
		return "", fmt.Errorf("ошибка генерации токена: %v", err)
	}
	return token, nil
}
