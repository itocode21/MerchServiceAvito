package services

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/itocode21/MerchServiceAvito/internal/repositories"
)

type ItemService struct {
	itemRepo *repositories.ItemRepository
	userRepo *repositories.UserRepository
}

func NewItemService(itemRepo *repositories.ItemRepository, userRepo *repositories.UserRepository) *ItemService {
	return &ItemService{itemRepo: itemRepo, userRepo: userRepo}
}

func (s *ItemService) BuyItem(username, itemName string) error {

	user, err := s.userRepo.GetUserByUsername(username)
	if err != nil {
		return fmt.Errorf("ошибка получения пользователя: %v", err)
	}
	if user == nil {
		return fmt.Errorf("пользователь %s не найден", username)
	}

	item, err := s.itemRepo.GetItemByName(itemName)
	if err != nil {
		return fmt.Errorf("ошибка получения предмета: %v", err)
	}
	if item == nil {
		return fmt.Errorf("предмет %s не найден", itemName)
	}

	if user.Coins < item.Price {
		return fmt.Errorf("недостаточно монет: %d < %d", user.Coins, item.Price)
	}

	tx, err := s.userRepo.DB.Begin()
	if err != nil {
		return fmt.Errorf("ошибка начала транзакции: %v", err)
	}
	defer tx.Rollback()

	user.Coins -= item.Price
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
