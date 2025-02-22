package database

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

func NewDB() (*sql.DB, error) {
	connStr := "postgresql://user:password@localhost:5432/avito_shop?sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}
	db.SetMaxOpenConns(50)
	return db, nil
}
