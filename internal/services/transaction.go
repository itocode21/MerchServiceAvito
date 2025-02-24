package services

import (
	"fmt"

	"github.com/itocode21/MerchServiceAvito/internal/models"
	"github.com/itocode21/MerchServiceAvito/internal/repositories"
)

type TransactionService struct {
	userRepo  *repositories.UserRepository
	transRepo *repositories.TransactionRepository
}

func NewTransactionService(userRepo *repositories.UserRepository, transRepo *repositories.TransactionRepository) *TransactionService {
	return &TransactionService{userRepo: userRepo, transRepo: transRepo}
}

func (s *TransactionService) SendCoins(fromUsername, toUsername string, amount int) error {
	if amount <= 0 {
		return fmt.Errorf("сумма должна быть положительной")
	}

	fromUser, err := s.userRepo.GetUserByUsername(fromUsername)
	if err != nil {
		return fmt.Errorf("ошибка получения отправителя: %v", err)
	}
	if fromUser == nil {
		return fmt.Errorf("отправитель %s не найден", fromUsername)
	}

	toUser, err := s.userRepo.GetUserByUsername(toUsername)
	if err != nil {
		return fmt.Errorf("ошибка получения получателя: %v", err)
	}
	if toUser == nil {
		return fmt.Errorf("получатель %s не найден", toUsername)
	}

	if fromUser.Coins < amount {
		return fmt.Errorf("недостаточно монет у %s: %d < %d", fromUsername, fromUser.Coins, amount)
	}

	tx, err := s.userRepo.DB.Begin()
	if err != nil {
		return fmt.Errorf("ошибка начала транзакции: %v", err)
	}
	defer tx.Rollback()

	fromUser.Coins -= amount
	if err := s.userRepo.UpdateUserBalanceTx(tx, fromUser); err != nil {
		return fmt.Errorf("ошибка обновления баланса отправителя: %v", err)
	}

	toUser.Coins += amount
	if err := s.userRepo.UpdateUserBalanceTx(tx, toUser); err != nil {
		return fmt.Errorf("ошибка обновления баланса получателя: %v", err)
	}

	transaction := &models.Transaction{
		FromUserID: fromUser.ID,
		ToUserID:   toUser.ID,
		Amount:     amount,
	}
	if err := s.transRepo.CreateTransaction(tx, transaction); err != nil {
		return fmt.Errorf("ошибка записи транзакции: %v", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("ошибка фиксации транзакции: %v", err)
	}

	return nil
}

func (s *TransactionService) GetUserTransactions(userID int) ([]models.Transaction, error) {
	return s.transRepo.GetUserTransactions(userID)
}
