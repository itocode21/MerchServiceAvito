package repositories

import (
	"database/sql"
	"fmt"

	"github.com/itocode21/MerchServiceAvito/internal/models"
)

type InventoryRepository struct {
	db *sql.DB
}

func NewInventoryRepository(db *sql.DB) *InventoryRepository {
	return &InventoryRepository{db: db}
}

func (r *InventoryRepository) GetUserInventory(userID string) ([]models.Inventory, error) {
	query := "SELECT id, user_id, item_id, quantity FROM inventory WHERE user_id = $1"
	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("error to get user inventory: %v", err)
	}
	defer rows.Close()

	var inventory []models.Inventory
	for rows.Next() {
		var item models.Inventory
		if err := rows.Scan(&item.ID, &item.UserID, &item.ItemID, &item.Quantity); err != nil {
			return nil, err
		}
		inventory = append(inventory, item)
	}
	return inventory, nil
}

func (r *InventoryRepository) AddItemToInventory(userID string, itemID int, quantity int) error {
	query := `
		INSERT INTO inventory (user_id, item_id, quantity) 
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, item_id) DO UPDATE 
		SET quantity = inventory.quantity + $3
	`
	_, err := r.db.Exec(query, userID, itemID, quantity)
	if err != nil {
		return fmt.Errorf("failed to add item to inventory: %v", err)
	}
	return nil
}
