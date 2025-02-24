package repositories

import (
	"database/sql"
	"fmt"

	"github.com/itocode21/MerchServiceAvito/internal/models"
)

type UserRepository struct {
	db *sql.DB
	DB *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db, DB: db}
}

func (r *UserRepository) GetUserByUsername(username string) (*models.User, error) {
	var user models.User
	query := "SELECT id, username, password_hash, coins FROM users WHERE username = $1"
	err := r.db.QueryRow(query, username).Scan(&user.ID, &user.Username, &user.PasswordHash, &user.Coins)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user by username: %v", err)
	}
	return &user, nil
}
func (r *UserRepository) CreateUser(user *models.User) error {
	query := `
        INSERT INTO users (username, password_hash, coins) 
        VALUES ($1, $2, $3) 
        RETURNING id, created_at
    `
	err := r.db.QueryRow(query, user.Username, user.PasswordHash, user.Coins).
		Scan(&user.ID, &user.CreatedAt)
	if err != nil {
		return fmt.Errorf("ошибка создания пользователя: %v", err)
	}
	return nil
}

func (r *UserRepository) UpdateUserBalance(user *models.User) error {
	query := "UPDATE users SET coins = $1 WHERE id = $2"
	_, err := r.db.Exec(query, user.Coins, user.ID)
	if err != nil {
		return fmt.Errorf("failed to update coins balance: %v", err)
	}
	return nil
}

func (r *UserRepository) UpdateUserBalanceTx(tx *sql.Tx, user *models.User) error {
	query := "UPDATE users SET coins = $1 WHERE id = $2"
	_, err := tx.Exec(query, user.Coins, user.ID)
	if err != nil {
		return fmt.Errorf("failed to update coins balance: %v", err)
	}
	return nil
}
