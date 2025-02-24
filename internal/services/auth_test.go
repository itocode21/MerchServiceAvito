package services

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/itocode21/MerchServiceAvito/internal/auth"
	"github.com/itocode21/MerchServiceAvito/internal/repositories"
	"golang.org/x/crypto/bcrypt"
)

func TestAuthenticate(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Ошибка создания мока: %v", err)
	}
	defer db.Close()

	userRepo := repositories.NewUserRepository(db)
	service := NewAuthService(userRepo)

	// Устанавливаем тестовый JWT-секрет перед тестами
	auth.SetJWTSecret([]byte("test_secret_key"))

	// Хешируем тестовый пароль
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("12345"), bcrypt.DefaultCost)

	tests := []struct {
		name      string
		username  string
		password  string
		setupMock func()
		wantErr   bool
		errMsg    string
	}{
		{
			name:     "Успешная аутентификация",
			username: "user1",
			password: "12345",
			setupMock: func() {
				mock.ExpectQuery("SELECT id, username, password_hash, coins FROM users WHERE username = \\$1").
					WithArgs("user1").
					WillReturnRows(sqlmock.NewRows([]string{"id", "username", "password_hash", "coins"}).
						AddRow(1, "user1", string(hashedPassword), 1000))
			},
			wantErr: false,
		},
		{
			name:     "Неверный пароль",
			username: "user1",
			password: "wrongpass",
			setupMock: func() {
				mock.ExpectQuery("SELECT id, username, password_hash, coins FROM users WHERE username = \\$1").
					WithArgs("user1").
					WillReturnRows(sqlmock.NewRows([]string{"id", "username", "password_hash", "coins"}).
						AddRow(1, "user1", string(hashedPassword), 1000))
			},
			wantErr: true,
			errMsg:  "неверный пароль",
		},
		{
			name:     "Пользователь не найден",
			username: "unknown",
			password: "12345",
			setupMock: func() {
				mock.ExpectQuery("SELECT id, username, password_hash, coins FROM users WHERE username = \\$1").
					WithArgs("unknown").
					WillReturnRows(sqlmock.NewRows([]string{"id", "username", "password_hash", "coins"}))
			},
			wantErr: true,
			errMsg:  "пользователь не найден",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()
			token, err := service.Authenticate(tt.username, tt.password)
			if tt.wantErr {
				if err == nil || err.Error() != tt.errMsg {
					t.Errorf("Authenticate() error = %v, wantErr %v, errMsg %q", err, tt.wantErr, tt.errMsg)
				}
				if token != "" {
					t.Errorf("Authenticate() token = %q, want empty", token)
				}
			} else if err != nil {
				t.Errorf("Authenticate() error = %v, want nil", err)
			} else if token == "" {
				t.Errorf("Authenticate() token is empty, want non-empty")
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("Не все ожидания мока выполнены: %v", err)
			}
		})
	}
}
