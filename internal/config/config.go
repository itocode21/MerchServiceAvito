package config

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/go-redis/redis/v8"
	"github.com/itocode21/MerchServiceAvito/internal/database"
	"github.com/joho/godotenv"
)

type Config struct {
	DB        *sql.DB
	JWTSecret []byte
	Redis     *redis.Client
}

func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		log.Printf("Не удалось загрузить .env: %v (будут использованы переменные окружения)", err)
	}

	// Подключение к базе
	db, err := database.NewDB()
	if err != nil {
		log.Printf("Ошибка подключения к БД: %v", err)
		return nil, err
	}

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}
	redisPassword := os.Getenv("REDIS_PASSWORD")
	redisClient := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       0,
	})
	if _, err := redisClient.Ping(context.Background()).Result(); err != nil {
		log.Printf("Ошибка подключения к Redis: %v", err)
		return nil, err
	}
	log.Println("Соединение с Redis успешно установлено")

	jwtSecret := []byte(os.Getenv("JWT_SECRET"))
	if len(jwtSecret) == 0 {
		log.Printf("JWT_SECRET не указан")
		return nil, fmt.Errorf("JWT_SECRET не указан")
	}

	return &Config{
		DB:        db,
		JWTSecret: jwtSecret,
		Redis:     redisClient,
	}, nil
}

func (c *Config) Close() {
	if err := c.DB.Close(); err != nil {
		log.Printf("Ошибка закрытия БД: %v", err)
	}
	if err := c.Redis.Close(); err != nil {
		log.Printf("Ошибка закрытия Redis: %v", err)
	}
}
