package services

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/itocode21/MerchServiceAvito/internal/repositories"
)

func TestBuyItem(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Ошибка создания мока: %v", err)
	}
	defer db.Close()

	userRepo := repositories.NewUserRepository(db)
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
				mock.ExpectQuery("SELECT id, username, password_hash, coins FROM users WHERE username = \\$1").
					WithArgs("user1").
					WillReturnRows(sqlmock.NewRows([]string{"id", "username", "password_hash", "coins"}).
						AddRow(1, "user1", "hash", 1000))

				mock.ExpectQuery("SELECT id, name, price FROM items WHERE name = \\$1").
					WithArgs("t-shirt").
					WillReturnRows(sqlmock.NewRows([]string{"id", "name", "price"}).
						AddRow(1, "t-shirt", 80))

				mock.ExpectBegin()
				mock.ExpectExec("UPDATE users SET coins = \\$1 WHERE id = \\$2").
					WithArgs(920, 1).
					WillReturnResult(sqlmock.NewResult(1, 1))
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
				mock.ExpectQuery("SELECT id, username, password_hash, coins FROM users WHERE username = \\$1").
					WithArgs("user1").
					WillReturnRows(sqlmock.NewRows([]string{"id", "username", "password_hash", "coins"}).
						AddRow(1, "user1", "hash", 200))
				mock.ExpectQuery("SELECT id, name, price FROM items WHERE name = \\$1").
					WithArgs("hoody").
					WillReturnRows(sqlmock.NewRows([]string{"id", "name", "price"}).
						AddRow(2, "hoody", 300))
			},
			wantErr: true,
			errMsg:  "недостаточно монет: 200 < 300",
		},
		{
			name:     "Предмет не найден",
			username: "user1",
			itemName: "nonexistent",
			setupMock: func() {
				mock.ExpectQuery("SELECT id, username, password_hash, coins FROM users WHERE username = \\$1").
					WithArgs("user1").
					WillReturnRows(sqlmock.NewRows([]string{"id", "username", "password_hash", "coins"}).
						AddRow(1, "user1", "hash", 1000))
				mock.ExpectQuery("SELECT id, name, price FROM items WHERE name = \\$1").
					WithArgs("nonexistent").
					WillReturnRows(sqlmock.NewRows([]string{"id", "name", "price"})) // Пустой результат
			},
			wantErr: true,
			errMsg:  "предмет nonexistent не найден",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()
			err := service.BuyItem(tt.username, tt.itemName)
			if tt.wantErr {
				if err == nil || err.Error() != tt.errMsg {
					t.Errorf("BuyItem() error = %v, wantErr %v, errMsg %q", err, tt.wantErr, tt.errMsg)
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
