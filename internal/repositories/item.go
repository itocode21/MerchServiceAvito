package repositories

import (
	"database/sql"
	"fmt"

	"github.com/itocode21/MerchServiceAvito/internal/models"
)

type ItemRepository struct {
	db *sql.DB
}

func NewItemRepository(db *sql.DB) *ItemRepository {
	return &ItemRepository{db: db}
}

func (r *ItemRepository) GetAllItems() ([]models.Item, error) {
	rows, err := r.db.Query("SELECT id, name, price FROM items")
	if err != nil {
		return nil, fmt.Errorf("failed to get items: %v", err)
	}
	defer rows.Close()

	var items []models.Item
	for rows.Next() {
		var item models.Item
		if err := rows.Scan(&item.ID, &item.Name, &item.Price); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

func (r *ItemRepository) GetItemByName(name string) (*models.Item, error) {
	var item models.Item
	query := "SELECT id, name, price FROM items WHERE name = $1"
	err := r.db.QueryRow(query, name).Scan(&item.ID, &item.Name, &item.Price)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get item by name: %v", err)
	}
	return &item, nil
}
