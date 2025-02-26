package services

import (
	"database/sql"
	"fmt"

	"github.com/itocode21/MerchServiceAvito/internal/models"
	"github.com/itocode21/MerchServiceAvito/internal/repositories"
)

type TransactionService struct {
	userRepo  *repositories.UserRepository
	transRepo *repositories.TransactionRepository
	db        *sql.DB
}

func NewTransactionService(userRepo *repositories.UserRepository, transRepo *repositories.TransactionRepository) *TransactionService {
	return &TransactionService{
		userRepo:  userRepo,
		transRepo: transRepo,
		db:        userRepo.DB,
	}
}

func (s *TransactionService) SendCoins(fromUsername, toUsername string, amount int) error {
	if amount <= 0 {
		return fmt.Errorf("сумма должна быть положительной")
	}

	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("ошибка начала транзакции: %v", err)
	}
	defer tx.Rollback()

	var fromUserID, fromUserCoins int
	err = tx.QueryRow("SELECT id, coins FROM users WHERE username = $1 FOR UPDATE", fromUsername).
		Scan(&fromUserID, &fromUserCoins)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("отправитель %s не найден", fromUsername)
		}
		return fmt.Errorf("ошибка блокировки отправителя: %v", err)
	}

	var toUserID, toUserCoins int
	err = tx.QueryRow("SELECT id, coins FROM users WHERE username = $1 FOR UPDATE", toUsername).
		Scan(&toUserID, &toUserCoins)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("получатель %s не найден", toUsername)
		}
		return fmt.Errorf("ошибка блокировки получателя: %v", err)
	}

	if fromUserCoins < amount {
		return fmt.Errorf("недостаточно монет у %s: %d < %d", fromUsername, fromUserCoins, amount)
	}

	fromUser := &models.User{ID: fromUserID, Coins: fromUserCoins - amount}
	toUser := &models.User{ID: toUserID, Coins: toUserCoins + amount}

	if err := s.userRepo.UpdateUserBalanceTx(tx, fromUser); err != nil {
		return fmt.Errorf("ошибка обновления баланса отправителя: %v", err)
	}
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
