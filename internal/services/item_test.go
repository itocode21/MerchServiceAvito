package services

import (
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-redis/redis/v8"
	"github.com/itocode21/MerchServiceAvito/internal/config"
	"github.com/itocode21/MerchServiceAvito/internal/repositories"
)

func TestBuyItem(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Ошибка создания мока: %v", err)
	}
	defer db.Close()

	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	//Config для тестов
	cfg := &config.Config{
		DB:        db,
		JWTSecret: []byte("test_secret_key"),
		Redis:     redisClient,
	}

	userRepo := repositories.NewUserRepository(cfg)
	itemRepo := repositories.NewItemRepository(db)
	service := NewItemService(itemRepo, userRepo)

	tests := []struct {
		name      string
		username  string
		itemName  string
		setupMock func()
		wantErr   bool
		errMsg    string
	}{
		{
			name:     "Успешная покупка",
			username: "user1",
			itemName: "t-shirt",
			setupMock: func() {
				// Мокаем SELECT FOR UPDATE
				mock.ExpectBegin()
				mock.ExpectQuery("SELECT id, coins FROM users WHERE username = \\$1 FOR UPDATE").
					WithArgs("user1").
					WillReturnRows(sqlmock.NewRows([]string{"id", "coins"}).
						AddRow(1, 1000))

				// Мокаем GetItemByName
				mock.ExpectQuery("SELECT id, name, price FROM items WHERE name = \\$1").
					WithArgs("t-shirt").
					WillReturnRows(sqlmock.NewRows([]string{"id", "name", "price"}).
						AddRow(1, "t-shirt", 80))

				// Мокаем UpdateUserBalanceTx
				mock.ExpectExec("UPDATE users SET coins = \\$1 WHERE id = \\$2").
					WithArgs(920, 1).
					WillReturnResult(sqlmock.NewResult(1, 1))

				// Мокаем AddToInventory
				mock.ExpectExec("INSERT INTO inventory \\(user_id, item_id, quantity\\) VALUES \\(\\$1, \\$2, 1\\) ON CONFLICT \\(user_id, item_id\\) DO UPDATE SET quantity = inventory.quantity \\+ 1").
					WithArgs(1, 1).
					WillReturnResult(sqlmock.NewResult(1, 1))

				mock.ExpectCommit()
			},
			wantErr: false,
		},
		{
			name:     "Недостаточно монет",
			username: "user1",
			itemName: "hoody",
			setupMock: func() {
				mock.ExpectBegin()
				mock.ExpectQuery("SELECT id, coins FROM users WHERE username = \\$1 FOR UPDATE").
					WithArgs("user1").
					WillReturnRows(sqlmock.NewRows([]string{"id", "coins"}).
						AddRow(1, 200))

				mock.ExpectQuery("SELECT id, name, price FROM items WHERE name = \\$1").
					WithArgs("hoody").
					WillReturnRows(sqlmock.NewRows([]string{"id", "name", "price"}).
						AddRow(2, "hoody", 300))

				mock.ExpectRollback() // Транзакция откатывается из-за ошибки
			},
			wantErr: true,
			errMsg:  "недостаточно монет: 200 < 300",
		},
		{
			name:     "Предмет не найден",
			username: "user1",
			itemName: "nonexistent",
			setupMock: func() {
				mock.ExpectBegin()
				mock.ExpectQuery("SELECT id, coins FROM users WHERE username = \\$1 FOR UPDATE").
					WithArgs("user1").
					WillReturnRows(sqlmock.NewRows([]string{"id", "coins"}).
						AddRow(1, 1000))

				mock.ExpectQuery("SELECT id, name, price FROM items WHERE name = \\$1").
					WithArgs("nonexistent").
					WillReturnRows(sqlmock.NewRows([]string{"id", "name", "price"}))

				mock.ExpectRollback()
			},
			wantErr: true,
			errMsg:  "предмет nonexistent не найден",
		},
		{
			name:     "Пользователь не найден",
			username: "user999",
			itemName: "t-shirt",
			setupMock: func() {
				mock.ExpectBegin()
				mock.ExpectQuery("SELECT id, coins FROM users WHERE username = \\$1 FOR UPDATE").
					WithArgs("user999").
					WillReturnError(sql.ErrNoRows)

				mock.ExpectRollback()
			},
			wantErr: true,
			errMsg:  "пользователь user999 не найден",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()
			err := service.BuyItem(tt.username, tt.itemName)
			if tt.wantErr {
				if err == nil {
					t.Errorf("BuyItem() error = nil, want error %q", tt.errMsg)
				} else if err.Error() != tt.errMsg {
					t.Errorf("BuyItem() error = %v, want %q", err, tt.errMsg)
				}
			} else if err != nil {
				t.Errorf("BuyItem() error = %v, want nil", err)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("Не все ожидания мока выполнены: %v", err)
			}
		})
	}
}
