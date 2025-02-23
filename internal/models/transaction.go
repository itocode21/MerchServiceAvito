package models

import "time"

type Transaction struct {
	ID         int       `json:"id"`
	FromUserID string    `json:"from_user_id"`
	ToUserID   string    `json:"to_user_id"`
	Amount     int       `json:"amount"`
	CreatedAt  time.Time `json:"created_at"`
}
