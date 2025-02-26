package repositories

import (
	"database/sql"
	"fmt"

	"github.com/itocode21/MerchServiceAvito/internal/config"
	"github.com/itocode21/MerchServiceAvito/internal/models"
)

type UserRepository struct {
	db     *sql.DB
	DB     *sql.DB
	Config *config.Config
}

func NewUserRepository(cfg *config.Config) *UserRepository {
	return &UserRepository{
		db:     cfg.DB,
		DB:     cfg.DB,
		Config: cfg,
	}
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

func (r *UserRepository) GetUserInfo(username string) (*models.UserInfo, error) {
	query := `
        SELECT 
            u.coins,
            COALESCE(
                json_agg(
                    json_build_object(
                        'type', i.name,
                        'quantity', inv.quantity
                    )
                ) FILTER (WHERE inv.user_id IS NOT NULL),
                '[]'::json
            ) AS inventory,
            COALESCE(
                json_agg(
                    json_build_object(
                        'fromUser', t.from_user_id,
                        'amount', t.amount
                    )
                ) FILTER (WHERE t.to_user_id = u.id),
                '[]'::json
            ) AS received,
            COALESCE(
                json_agg(
                    json_build_object(
                        'toUser', t.to_user_id,
                        'amount', t.amount
                    )
                ) FILTER (WHERE t.from_user_id = u.id),
                '[]'::json
            ) AS sent
        FROM users u
        LEFT JOIN inventory inv ON inv.user_id = u.id
        LEFT JOIN items i ON i.id = inv.item_id
        LEFT JOIN transactions t ON t.from_user_id = u.id OR t.to_user_id = u.id
        WHERE u.username = $1
        GROUP BY u.id, u.coins
    `

	var info models.UserInfo
	err := r.db.QueryRow(query, username).Scan(&info.Coins, &info.InventoryJSON, &info.ReceivedJSON, &info.SentJSON)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user info: %v", err)
	}
	return &info, nil
}
