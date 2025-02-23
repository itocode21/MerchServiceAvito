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

func (r *TransactionRepository) AddTransaction(fromUserID, toUserID string, amount int) error {
	query := `
		INSERT INTO transactions (from_user_id, to_user_id, amount) 
		VALUES ($1, $2, $3)
	`
	_, err := r.db.Exec(query, fromUserID, toUserID, amount)
	if err != nil {
		return fmt.Errorf("failed to add transaction: %v", err)
	}
	return nil
}

func (r *TransactionRepository) GetUserTransactions(userID string) ([]models.Transaction, error) {
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
		var transaction models.Transaction
		if err := rows.Scan(&transaction.ID, &transaction.FromUserID, &transaction.ToUserID, &transaction.Amount, &transaction.CreatedAt); err != nil {
			return nil, err
		}
		transactions = append(transactions, transaction)
	}
	return transactions, nil
}
