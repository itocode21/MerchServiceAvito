package services

import (
	"fmt"

	"github.com/itocode21/MerchServiceAvito/internal/models"
	"github.com/itocode21/MerchServiceAvito/internal/repositories"
)

type ItemService struct {
	itemRepo *repositories.ItemRepository
}

func NewItemService(itemRepo *repositories.ItemRepository) *ItemService {
	return &ItemService{itemRepo: itemRepo}
}

func (s *ItemService) GetAllItems() ([]models.Item, error) {
	items, err := s.itemRepo.GetAllItems()
	if err != nil {
		return nil, fmt.Errorf("ошибка при получении списка товаров: %v", err)
	}
	return items, nil
}

func (s *ItemService) GetItemByName(name string) (*models.Item, error) {
	item, err := s.itemRepo.GetItemByName(name)
	if err != nil {
		return nil, fmt.Errorf("ошибка при получении товара: %v", err)
	}
	if item == nil {
		return nil, fmt.Errorf("товар с названием '%s' не найден", name)
	}
	return item, nil
}
