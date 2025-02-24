package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func NewDB() (*sql.DB, error) {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Ошибка загрузки .env файла: %v", err)
	}

	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_SSLMODE"),
	)
	log.Printf("Подключаемся к базе: %s", connStr)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("Ошибка подключения к базе данных: %v", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("Не удалось проверить соединение: %v", err)
	}
	log.Println("Соединение с базой успешно установлено")

	if err := applyMigrations(db); err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(50)
	return db, nil
}
func applyMigrations(db *sql.DB) error {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		log.Printf("Ошибка настройки драйвера миграций: %v", err)
		return fmt.Errorf("Ошибка настройки драйвера миграций: %v", err)
	}

	migrationsPath := "internal/database/migrations"
	log.Printf("Путь к миграциям: %s", migrationsPath)

	m, err := migrate.NewWithDatabaseInstance(
		"file://"+migrationsPath,
		"postgres", driver)
	if err != nil {
		log.Printf("Ошибка инициализации миграций: %v", err)
		return fmt.Errorf("Ошибка инициализации миграций: %v", err)
	}

	err = m.Up()
	if err != nil {
		if err == migrate.ErrNoChange {
			log.Println("Миграции не требуются, изменений нет")
		} else {
			log.Printf("Ошибка применения миграций: %v", err)
			return fmt.Errorf("Ошибка применения миграций: %v", err)
		}
	} else {
		log.Println("Миграции успешно применены")
	}

	var exists bool
	err = db.QueryRow("SELECT EXISTS (SELECT FROM pg_tables WHERE schemaname = 'public' AND tablename = 'users')").Scan(&exists)
	if err != nil {
		log.Printf("Ошибка проверки таблицы users: %v", err)
		return err
	}
	if !exists {
		log.Println("Таблица users не создана после миграций!")
	} else {
		log.Println("Таблица users успешно создана")
	}

	return nil
}
