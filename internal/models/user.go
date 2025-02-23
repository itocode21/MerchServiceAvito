package models

import "time"

type User struct {
	ID           string    `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"`
	Coins        int       `json:"coins"`
	CreatedAt    time.Time `json:"created_at"`
}
