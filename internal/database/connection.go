package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func NewDB() (*sql.DB, error) {
	err := godotenv.Load()
	if err != nil {
		log.Printf("Ошибка загрузки .env файла: %v", err)
	}

	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"), os.Getenv("DB_NAME"), os.Getenv("DB_SSLMODE"),
	)
	log.Printf("Подключаемся к базе: %s", connStr)

	var db *sql.DB
	for i := 0; i < 10; i++ {
		db, err = sql.Open("postgres", connStr)
		if err != nil {
			log.Printf("Ошибка открытия соединения: %v", err)
			time.Sleep(2 * time.Second)
			continue
		}

		if err = db.Ping(); err != nil {
			log.Printf("Не удалось проверить соединение: %v", err)
			db.Close()
			time.Sleep(2 * time.Second)
			continue
		}

		log.Println("Соединение с базой успешно установлено")
		log.Printf("Активных соединений после Ping: %d, максимум: %d", db.Stats().OpenConnections, db.Stats().MaxOpenConnections)

		if err := applyMigrations(db); err != nil {
			return nil, err
		}

		db.SetMaxOpenConns(1000)
		db.SetMaxIdleConns(500)

		log.Printf("Активных соединений после настройки пула: %d, максимум: %d", db.Stats().OpenConnections, db.Stats().MaxOpenConnections)

		return db, nil
	}

	return nil, fmt.Errorf("не удалось подключиться к базе после 10 попыток: %v", err)
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

func ResetDB(db *sql.DB) error {
	_, err := db.Exec("TRUNCATE TABLE users, transactions, inventory RESTART IDENTITY")
	if err != nil {
		log.Printf("Ошибка очистки базы данных: %v", err)
		return fmt.Errorf("ошибка очистки базы данных: %v", err)
	}
	log.Println("База данных успешно очищена")
	return nil
}
