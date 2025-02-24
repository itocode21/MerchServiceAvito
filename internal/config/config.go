package config

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/itocode21/MerchServiceAvito/internal/database"
	"github.com/joho/godotenv"
)

type Config struct {
	DB        *sql.DB
	JWTSecret []byte
}

func Load() (*Config, error) {

	if err := godotenv.Load(); err != nil {
		log.Printf("Ошибка загрузки .env: %v", err)
		return nil, err
	}

	db, err := database.NewDB()
	if err != nil {
		log.Printf("Ошибка подключения к БД: %v", err)
		return nil, err
	}

	jwtSecret := []byte(os.Getenv("JWT_SECRET"))
	if len(jwtSecret) == 0 {
		log.Printf("JWT_SECRET не указан в .env")
		return nil, fmt.Errorf("JWT_SECRET не указан в .env")
	}

	return &Config{
		DB:        db,
		JWTSecret: jwtSecret,
	}, nil
}

func (c *Config) Close() {
	if err := c.DB.Close(); err != nil {
		log.Printf("Ошибка закрытия БД: %v", err)
	}
}
