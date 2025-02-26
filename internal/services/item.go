package services

import (
	"database/sql"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/itocode21/MerchServiceAvito/internal/models"
	"github.com/itocode21/MerchServiceAvito/internal/repositories"
)

type ItemService struct {
	itemRepo *repositories.ItemRepository
	userRepo *repositories.UserRepository
	db       *sql.DB
}

func NewItemService(itemRepo *repositories.ItemRepository, userRepo *repositories.UserRepository) *ItemService {
	return &ItemService{
		itemRepo: itemRepo,
		userRepo: userRepo,
		db:       userRepo.DB,
	}
}

func (s *ItemService) BuyItem(username, itemName string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("ошибка начала транзакции: %v", err)
	}
	defer tx.Rollback()

	var userID int
	var userCoins int
	err = tx.QueryRow("SELECT id, coins FROM users WHERE username = $1 FOR UPDATE", username).
		Scan(&userID, &userCoins)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("пользователь %s не найден", username)
		}
		return fmt.Errorf("ошибка блокировки пользователя: %v", err)
	}

	item, err := s.itemRepo.GetItemByName(itemName)
	if err != nil {
		return fmt.Errorf("ошибка получения предмета: %v", err)
	}
	if item == nil {
		return fmt.Errorf("предмет %s не найден", itemName)
	}

	if userCoins < item.Price {
		return fmt.Errorf("недостаточно монет: %d < %d", userCoins, item.Price)
	}

	user := &models.User{ID: userID, Coins: userCoins - item.Price}
	if err := s.userRepo.UpdateUserBalanceTx(tx, user); err != nil {
		return fmt.Errorf("ошибка обновления баланса: %v", err)
	}

	if err := s.itemRepo.AddToInventory(tx, user.ID, item.ID); err != nil {
		return fmt.Errorf("ошибка добавления в инвентарь: %v", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("ошибка фиксации транзакции: %v", err)
	}

	return nil
}

func (s *ItemService) GetUserInventory(userID int) ([]gin.H, error) {
	return s.itemRepo.GetUserInventory(userID)
}
