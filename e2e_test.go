package test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/itocode21/MerchServiceAvito/internal/auth"
	"github.com/itocode21/MerchServiceAvito/internal/config"
	"github.com/itocode21/MerchServiceAvito/internal/handlers"
	"github.com/itocode21/MerchServiceAvito/internal/middleware"
	"github.com/itocode21/MerchServiceAvito/internal/repositories"
	"github.com/itocode21/MerchServiceAvito/internal/services"
)

// setupTest инициализирует окружение для E2E-тестов
func setupTest(t *testing.T) (*gin.Engine, *sql.DB, *redis.Client, func()) {
	projectDir := `d:\project\MerchServiceAvito`

	// Запускаем docker-compose
	cmd := exec.Command("docker-compose", "up", "-d")
	cmd.Dir = projectDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Не удалось запустить docker-compose: %v\nВывод: %s", err, string(output))
	}

	// Даём время контейнерам подняться
	time.Sleep(5 * time.Second)

	// Подключаемся к PostgreSQL
	db, err := sql.Open("postgres", "host=localhost port=5432 user=postgres password=root dbname=avito_shop sslmode=disable")
	if err != nil {
		t.Fatalf("Ошибка подключения к PostgreSQL: %v", err)
	}
	if err := db.Ping(); err != nil {
		t.Fatalf("Ошибка проверки PostgreSQL: %v", err)
	}

	// Подключаемся к Redis
	redisClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "your_redis_password", // Пароль из docker-compose.yml
	})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		t.Fatalf("Ошибка подключения к Redis: %v", err)
	}

	// Создаём конфигурацию
	cfg := &config.Config{
		DB:        db,
		Redis:     redisClient,
		JWTSecret: []byte("your_very_secure_secret_key_32_bytes_long"),
	}

	// Инициализируем репозитории и сервисы
	userRepo := repositories.NewUserRepository(cfg)
	itemRepo := repositories.NewItemRepository(cfg.DB)
	transRepo := repositories.NewTransactionRepository(cfg.DB)
	userService := services.NewUserService(userRepo)
	authService := services.NewAuthService(userRepo)
	itemService := services.NewItemService(itemRepo, userRepo)
	transService := services.NewTransactionService(userRepo, transRepo)
	auth.SetJWTSecret(cfg.JWTSecret)

	h := handlers.NewHandlers(cfg, authService, userService, itemService, transService)

	// Настраиваем маршруты
	r := gin.Default()
	r.POST("/api/register", h.Register)
	r.POST("/api/auth", h.Authenticate)
	protected := r.Group("/api").Use(middleware.JWTAuthMiddleware())
	protected.GET("/info", h.GetInfo)
	protected.POST("/sendCoin", h.SendCoin)
	protected.GET("/buy/:item", h.BuyItem)

	cleanup := func() {
		db.Exec("TRUNCATE TABLE transactions, inventory, users RESTART IDENTITY CASCADE")
		db.Close()
		redisClient.Close()
		cmd := exec.Command("docker-compose", "down")
		cmd.Dir = projectDir
		if err := cmd.Run(); err != nil {
			t.Logf("Не удалось остановить docker-compose: %v", err)
		}
	}

	return r, db, redisClient, cleanup
}

// TestE2EBuyItem проверяет сценарий покупки мерча
func TestE2EBuyItem(t *testing.T) {
	r, _, _, cleanup := setupTest(t)
	defer cleanup()

	// Регистрация пользователя
	registerReq := map[string]string{"username": "user1", "password": "12345"}
	registerBody, _ := json.Marshal(registerReq)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/register", bytes.NewBuffer(registerBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("Регистрация провалилась: %d, %s", w.Code, w.Body.String())
	}

	// Аутентификация
	authReq := map[string]string{"username": "user1", "password": "12345"}
	authBody, _ := json.Marshal(authReq)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/auth", bytes.NewBuffer(authBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("Аутентификация провалилась: %d, %s", w.Code, w.Body.String())
	}
	var authResp struct{ Token string }
	json.Unmarshal(w.Body.Bytes(), &authResp)

	// Покупка предмета
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/buy/t-shirt", nil)
	req.Header.Set("Authorization", "Bearer "+authResp.Token)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("Покупка провалилась: %d, %s", w.Code, w.Body.String())
	}

	// Проверка результата через /api/info
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/info", nil)
	req.Header.Set("Authorization", "Bearer "+authResp.Token)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("Получение информации провалилось: %d, %s", w.Code, w.Body.String())
	}

	var infoResp struct {
		Coins     int `json:"coins"`
		Inventory []struct {
			Type     string `json:"type"`
			Quantity int    `json:"quantity"`
		} `json:"inventory"`
		CoinHistory struct {
			Received []interface{} `json:"received"`
			Sent     []interface{} `json:"sent"`
		} `json:"coinHistory"`
	}
	json.Unmarshal(w.Body.Bytes(), &infoResp)
	if infoResp.Coins != 920 {
		t.Errorf("Ожидалось 920 монет, получено %d", infoResp.Coins)
	}
	if len(infoResp.Inventory) != 1 || infoResp.Inventory[0].Type != "t-shirt" || infoResp.Inventory[0].Quantity != 1 {
		t.Errorf("Ожидался t-shirt с количеством 1, получено %v", infoResp.Inventory)
	}
}

// TestE2ESendCoin проверяет сценарий передачи монеток другим сотрудникам
func TestE2ESendCoin(t *testing.T) {
	r, _, _, cleanup := setupTest(t)
	defer cleanup()

	// Регистрация двух пользователей
	for _, user := range []struct{ username, password string }{
		{"sender", "12345"},
		{"receiver", "12345"},
	} {
		reqBody, _ := json.Marshal(map[string]string{"username": user.username, "password": user.password})
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/register", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("Регистрация %s провалилась: %d, %s", user.username, w.Code, w.Body.String())
		}
	}

	// Аутентификация отправителя
	authReq := map[string]string{"username": "sender", "password": "12345"}
	authBody, _ := json.Marshal(authReq)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/auth", bytes.NewBuffer(authBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("Аутентификация провалилась: %d, %s", w.Code, w.Body.String())
	}
	var authResp struct{ Token string }
	json.Unmarshal(w.Body.Bytes(), &authResp)

	// Передача монет
	sendReq := map[string]interface{}{"toUser": "receiver", "amount": 100}
	sendBody, _ := json.Marshal(sendReq)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/sendCoin", bytes.NewBuffer(sendBody))
	req.Header.Set("Authorization", "Bearer "+authResp.Token)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("Передача монет провалилась: %d, %s", w.Code, w.Body.String())
	}

	// Проверка результата через /api/info для sender
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/info", nil)
	req.Header.Set("Authorization", "Bearer "+authResp.Token)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("Получение информации для sender провалилось: %d, %s", w.Code, w.Body.String())
	}

	var senderInfoResp struct {
		Coins     int `json:"coins"`
		Inventory []struct {
			Type     string `json:"type"`
			Quantity int    `json:"quantity"`
		} `json:"inventory"`
		CoinHistory struct {
			Received []interface{} `json:"received"`
			Sent     []struct {
				ToUser int `json:"toUser"`
				Amount int `json:"amount"`
			} `json:"sent"`
		} `json:"coinHistory"`
	}
	json.Unmarshal(w.Body.Bytes(), &senderInfoResp)
	if senderInfoResp.Coins != 900 {
		t.Errorf("Ожидалось 900 монет у sender, получено %d", senderInfoResp.Coins)
	}
	if len(senderInfoResp.CoinHistory.Sent) != 1 || senderInfoResp.CoinHistory.Sent[0].Amount != 100 || senderInfoResp.CoinHistory.Sent[0].ToUser != 2 {
		t.Errorf("Ожидалась одна отправленная транзакция на 100 монет пользователю с ID 2, получено %v", senderInfoResp.CoinHistory.Sent)
	}

	// Аутентификация получателя
	authReq = map[string]string{"username": "receiver", "password": "12345"}
	authBody, _ = json.Marshal(authReq)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/auth", bytes.NewBuffer(authBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("Аутентификация получателя провалилась: %d, %s", w.Code, w.Body.String())
	}
	json.Unmarshal(w.Body.Bytes(), &authResp)

	// Проверка результата через /api/info для receiver
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/info", nil)
	req.Header.Set("Authorization", "Bearer "+authResp.Token)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("Получение информации для receiver провалилось: %d, %s", w.Code, w.Body.String())
	}

	var receiverInfoResp struct {
		Coins     int `json:"coins"`
		Inventory []struct {
			Type     string `json:"type"`
			Quantity int    `json:"quantity"`
		} `json:"inventory"`
		CoinHistory struct {
			Received []struct {
				FromUser int `json:"fromUser"`
				Amount   int `json:"amount"`
			} `json:"received"`
			Sent []interface{} `json:"sent"`
		} `json:"coinHistory"`
	}
	json.Unmarshal(w.Body.Bytes(), &receiverInfoResp)
	if receiverInfoResp.Coins != 1100 {
		t.Errorf("Ожидалось 1100 монет у receiver, получено %d", receiverInfoResp.Coins)
	}
	if len(receiverInfoResp.CoinHistory.Received) != 1 || receiverInfoResp.CoinHistory.Received[0].Amount != 100 || receiverInfoResp.CoinHistory.Received[0].FromUser != 1 {
		t.Errorf("Ожидалась одна полученная транзакция на 100 монет от пользователя с ID 1, получено %v", receiverInfoResp.CoinHistory.Received)
	}
}
