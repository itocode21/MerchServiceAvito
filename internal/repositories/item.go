package repositories

import (
	"database/sql"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/itocode21/MerchServiceAvito/internal/models"
)

type ItemRepository struct {
	db *sql.DB
}

func NewItemRepository(db *sql.DB) *ItemRepository {
	return &ItemRepository{db: db}
}

func (r *ItemRepository) GetItemByName(name string) (*models.Item, error) {
	var item models.Item
	query := "SELECT id, name, price FROM items WHERE name = $1"
	err := r.db.QueryRow(query, name).Scan(&item.ID, &item.Name, &item.Price)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get item: %v", err)
	}
	return &item, nil
}

func (r *ItemRepository) AddToInventory(tx *sql.Tx, userID, itemID int) error {
	query := `
        INSERT INTO inventory (user_id, item_id, quantity)
        VALUES ($1, $2, 1)
        ON CONFLICT (user_id, item_id)
        DO UPDATE SET quantity = inventory.quantity + 1
    `
	_, err := tx.Exec(query, userID, itemID)
	if err != nil {
		return fmt.Errorf("failed to add to inventory: %v", err)
	}
	return nil
}

func (r *ItemRepository) GetUserInventory(userID int) ([]gin.H, error) {
	query := `
        SELECT i.name, inv.quantity
        FROM inventory inv
        JOIN items i ON i.id = inv.item_id
        WHERE inv.user_id = $1
    `
	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения инвентаря: %v", err)
	}
	defer rows.Close()

	var inventory []gin.H
	for rows.Next() {
		var name string
		var quantity int
		if err := rows.Scan(&name, &quantity); err != nil {
			return nil, fmt.Errorf("ошибка сканирования инвентаря: %v", err)
		}
		inventory = append(inventory, gin.H{"type": name, "quantity": quantity})
	}
	return inventory, nil
}
