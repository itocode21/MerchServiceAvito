package services

import (
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/itocode21/MerchServiceAvito/internal/repositories"
)

func TestSendCoins(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Ошибка создания мока: %v", err)
	}
	defer db.Close()

	userRepo := repositories.NewUserRepository(db)
	transRepo := repositories.NewTransactionRepository(db)
	service := NewTransactionService(userRepo, transRepo)

	tests := []struct {
		name         string
		fromUsername string
		toUsername   string
		amount       int
		setupMock    func()
		wantErr      bool
		errMsg       string
	}{
		{
			name:         "Успешная передача монет",
			fromUsername: "user1",
			toUsername:   "user2",
			amount:       100,
			setupMock: func() {
				mock.ExpectQuery("SELECT id, username, password_hash, coins FROM users WHERE username = \\$1").
					WithArgs("user1").
					WillReturnRows(sqlmock.NewRows([]string{"id", "username", "password_hash", "coins"}).
						AddRow(1, "user1", "hash", 1000))
				mock.ExpectQuery("SELECT id, username, password_hash, coins FROM users WHERE username = \\$1").
					WithArgs("user2").
					WillReturnRows(sqlmock.NewRows([]string{"id", "username", "password_hash", "coins"}).
						AddRow(2, "user2", "hash", 500))
				mock.ExpectBegin()
				mock.ExpectExec("UPDATE users SET coins = \\$1 WHERE id = \\$2").
					WithArgs(900, 1).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec("UPDATE users SET coins = \\$1 WHERE id = \\$2").
					WithArgs(600, 2).
					WillReturnResult(sqlmock.NewResult(1, 1))
				createdAt, _ := time.Parse(time.RFC3339, "2025-02-24T12:00:00Z")
				mock.ExpectQuery("INSERT INTO transactions \\(from_user_id, to_user_id, amount\\) VALUES \\(\\$1, \\$2, \\$3\\) RETURNING id, created_at").
					WithArgs(1, 2, 100).
					WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}).AddRow(1, createdAt))
				mock.ExpectCommit()
			},
			wantErr: false,
		},
		{
			name:         "Недостаточно монет",
			fromUsername: "user1",
			toUsername:   "user2",
			amount:       2000,
			setupMock: func() {
				mock.ExpectQuery("SELECT id, username, password_hash, coins FROM users WHERE username = \\$1").
					WithArgs("user1").
					WillReturnRows(sqlmock.NewRows([]string{"id", "username", "password_hash", "coins"}).
						AddRow(1, "user1", "hash", 1000))
				mock.ExpectQuery("SELECT id, username, password_hash, coins FROM users WHERE username = \\$1").
					WithArgs("user2").
					WillReturnRows(sqlmock.NewRows([]string{"id", "username", "password_hash", "coins"}).
						AddRow(2, "user2", "hash", 500))
			},
			wantErr: true,
			errMsg:  "недостаточно монет у user1: 1000 < 2000",
		},
		{
			name:         "Отправитель не найден",
			fromUsername: "unknown",
			toUsername:   "user2",
			amount:       100,
			setupMock: func() {
				mock.ExpectQuery("SELECT id, username, password_hash, coins FROM users WHERE username = \\$1").
					WithArgs("unknown").
					WillReturnRows(sqlmock.NewRows([]string{"id", "username", "password_hash", "coins"}))
			},
			wantErr: true,
			errMsg:  "отправитель unknown не найден",
		},
		{
			name:         "Получатель не найден",
			fromUsername: "user1",
			toUsername:   "unknown",
			amount:       100,
			setupMock: func() {
				mock.ExpectQuery("SELECT id, username, password_hash, coins FROM users WHERE username = \\$1").
					WithArgs("user1").
					WillReturnRows(sqlmock.NewRows([]string{"id", "username", "password_hash", "coins"}).
						AddRow(1, "user1", "hash", 1000))
				mock.ExpectQuery("SELECT id, username, password_hash, coins FROM users WHERE username = \\$1").
					WithArgs("unknown").
					WillReturnRows(sqlmock.NewRows([]string{"id", "username", "password_hash", "coins"}))
			},
			wantErr: true,
			errMsg:  "получатель unknown не найден",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()
			err := service.SendCoins(tt.fromUsername, tt.toUsername, tt.amount)
			if tt.wantErr {
				if err == nil || err.Error() != tt.errMsg {
					t.Errorf("SendCoins() error = %v, wantErr %v, errMsg %q", err, tt.wantErr, tt.errMsg)
				}
			} else if err != nil {
				t.Errorf("SendCoins() error = %v, want nil", err)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("Не все ожидания мока выполнены: %v", err)
			}
		})
	}
}
