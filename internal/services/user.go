package services

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/itocode21/MerchServiceAvito/internal/models"
	"github.com/itocode21/MerchServiceAvito/internal/repositories"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	userRepo   *repositories.UserRepository
	workerPool chan struct{} // Пул горутин
}

func NewUserService(userRepo *repositories.UserRepository) *UserService {
	return &UserService{
		userRepo:   userRepo,
		workerPool: make(chan struct{}, 100), //100 горутинами
	}
}

func (s *UserService) GetUserInfo(username string) (*models.UserInfo, error) {
	return s.userRepo.GetUserInfo(username)
}

func (s *UserService) GetUserByUsername(username string) (*models.User, error) {
	cacheKey := "user:" + username
	cached, err := s.userRepo.Config.Redis.Get(context.Background(), cacheKey).Result()
	if err == nil {
		var user models.User
		if err := json.Unmarshal([]byte(cached), &user); err == nil {
			return &user, nil
		}
	}

	user, err := s.userRepo.GetUserByUsername(username)
	if err != nil {
		return nil, fmt.Errorf("ошибка при получении пользователя: %v", err)
	}
	if user == nil {
		return nil, fmt.Errorf("пользователь не найден")
	}

	userJSON, _ := json.Marshal(user)
	s.userRepo.Config.Redis.Set(context.Background(), cacheKey, userJSON, 5*time.Minute)
	return user, nil
}

func (s *UserService) RegisterUser(username, password string) (*models.User, error) {
	cacheKey := "user:" + username
	cached, err := s.userRepo.Config.Redis.Get(context.Background(), cacheKey).Result()
	if err == nil {
		var user models.User
		if err := json.Unmarshal([]byte(cached), &user); err == nil {
			return &user, nil
		}
	}

	existsKey := "user_exists:" + username
	exists, err := s.userRepo.Config.Redis.Get(context.Background(), existsKey).Result()
	if err == nil && exists == "true" {
		return nil, fmt.Errorf("пользователь с таким именем уже существует")
	}

	var wg sync.WaitGroup
	var checkErr, hashErr error
	var existingUser *models.User
	var passwordHash []byte
	var user *models.User

	s.workerPool <- struct{}{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer func() { <-s.workerPool }()
		existingUser, checkErr = s.userRepo.GetUserByUsername(username)
	}()

	s.workerPool <- struct{}{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer func() { <-s.workerPool }()
		passwordHash, hashErr = bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	}()

	wg.Wait()
	if hashErr != nil {
		return nil, fmt.Errorf("ошибка хеширования пароля: %v", hashErr)
	}
	if checkErr != nil {
		return nil, checkErr
	}
	if existingUser != nil {
		s.userRepo.Config.Redis.Set(context.Background(), existsKey, "true", 5*time.Minute)
		userJSON, _ := json.Marshal(existingUser)
		s.userRepo.Config.Redis.Set(context.Background(), cacheKey, userJSON, 5*time.Minute)
		return nil, fmt.Errorf("пользователь с таким именем уже существует")
	}

	user = &models.User{
		Username:     username,
		PasswordHash: string(passwordHash),
		Coins:        1000,
	}

	// Синхронная вставка
	if err := s.userRepo.CreateUser(user); err != nil {
		return nil, fmt.Errorf("ошибка при создании пользователя: %v", err)
	}

	// хеш пароля в Redis
	hashCacheKey := "user_hash:" + username
	s.userRepo.Config.Redis.Set(context.Background(), hashCacheKey, user.PasswordHash, 5*time.Minute)

	userJSON, _ := json.Marshal(user)
	s.userRepo.Config.Redis.Set(context.Background(), cacheKey, userJSON, 5*time.Minute)
	s.userRepo.Config.Redis.Set(context.Background(), existsKey, "true", 5*time.Minute)
	return user, nil
}

func (s *UserService) UpdateUserBalance(username string, amount int) error {
	user, err := s.userRepo.GetUserByUsername(username)
	if err != nil {
		return err
	}
	if user == nil {
		return fmt.Errorf("пользователь не найден")
	}

	newBalance := user.Coins + amount
	if newBalance < 0 {
		return fmt.Errorf("недостаточно монет")
	}

	user.Coins = newBalance
	if err := s.userRepo.UpdateUserBalance(user); err != nil {
		return fmt.Errorf("ошибка при обновлении баланса: %v", err)
	}
	return nil
}
