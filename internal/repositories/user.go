package repositories

import (
	"database/sql"
	"fmt"

	"github.com/itocode21/MerchServiceAvito/internal/models"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) GetUserByUsername(username string) (*models.User, error) {
	var user models.User
	query := "SELECT id, username, coins FROM users WHERE username = $1"
	err := r.db.QueryRow(query, username).Scan(&user.ID, &user.Username, &user.Coins)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user by username: %v", err)
	}
	return &user, nil
}

func (r *UserRepository) CreateUser(user *models.User) error {
	query := "INSERT INTO users (username, password_hash, coins) VALUES ($1, $2, $3)"
	_, err := r.db.Exec(query, user.Username, user.PasswordHash, user.Coins)
	if err != nil {
		return fmt.Errorf("failed to create user: %v", err)
	}
	return nil
}
