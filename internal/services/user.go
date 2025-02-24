package services

import (
	"fmt"

	"github.com/itocode21/MerchServiceAvito/internal/models"
	"github.com/itocode21/MerchServiceAvito/internal/repositories"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	userRepo *repositories.UserRepository
}

func NewUserService(userRepo *repositories.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

func (s *UserService) GetUserByUsername(username string) (*models.User, error) {
	user, err := s.userRepo.GetUserByUsername(username)
	if err != nil {
		return nil, fmt.Errorf("ошибка при получении пользователя: %v", err)
	}
	if user == nil {
		return nil, fmt.Errorf("пользователь не найден")
	}
	return user, nil
}

func (s *UserService) RegisterUser(username, password string) (*models.User, error) {

	existingUser, err := s.userRepo.GetUserByUsername(username)
	if err != nil {
		return nil, err
	}
	if existingUser != nil {
		return nil, fmt.Errorf("пользователь с таким именем уже существует")
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("ошибка хеширования пароля: %v", err)
	}

	user := &models.User{
		Username:     username,
		PasswordHash: string(passwordHash),
		Coins:        1000, // Стартовый баланс
	}
	if err := s.userRepo.CreateUser(user); err != nil {
		return nil, fmt.Errorf("ошибка при создании пользователя: %v", err)
	}
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
