package repositories

import (
	"database/sql"
	"fmt"

	"github.com/itocode21/MerchServiceAvito/internal/models"
)

type TransactionRepository struct {
	db *sql.DB
}

func NewTransactionRepository(db *sql.DB) *TransactionRepository {
	return &TransactionRepository{db: db}
}

func (r *TransactionRepository) CreateTransaction(tx *sql.Tx, t *models.Transaction) error {
	query := `
        INSERT INTO transactions (from_user_id, to_user_id, amount)
        VALUES ($1, $2, $3)
        RETURNING id, created_at
    `
	err := tx.QueryRow(query, t.FromUserID, t.ToUserID, t.Amount).Scan(&t.ID, &t.CreatedAt)
	if err != nil {
		return fmt.Errorf("ошибка создания транзакции: %v", err)
	}
	return nil
}

func (r *TransactionRepository) GetUserTransactions(userID int) ([]models.Transaction, error) {
	query := `
        SELECT id, from_user_id, to_user_id, amount, created_at 
        FROM transactions 
        WHERE from_user_id = $1 OR to_user_id = $1
    `
	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user transactions: %v", err)
	}
	defer rows.Close()

	var transactions []models.Transaction
	for rows.Next() {
		var t models.Transaction
		if err := rows.Scan(&t.ID, &t.FromUserID, &t.ToUserID, &t.Amount, &t.CreatedAt); err != nil {
			return nil, fmt.Errorf("ошибка сканирования транзакции: %v", err)
		}
		transactions = append(transactions, t)
	}
	return transactions, nil
}
