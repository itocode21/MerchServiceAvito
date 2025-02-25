package integration

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/itocode21/MerchServiceAvito/internal/auth"
	"github.com/itocode21/MerchServiceAvito/internal/config"
	"github.com/itocode21/MerchServiceAvito/internal/handlers"
	"github.com/itocode21/MerchServiceAvito/internal/middleware"
	"github.com/itocode21/MerchServiceAvito/internal/repositories"
	"github.com/itocode21/MerchServiceAvito/internal/services"
)

func setupTest(t *testing.T) (*gin.Engine, *sql.DB, func()) {
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Ошибка загрузки конфигурации: %v", err)
	}

	userRepo := repositories.NewUserRepository(cfg.DB)
	itemRepo := repositories.NewItemRepository(cfg.DB)
	transRepo := repositories.NewTransactionRepository(cfg.DB)
	userService := services.NewUserService(userRepo)
	authService := services.NewAuthService(userRepo)
	itemService := services.NewItemService(itemRepo, userRepo)
	transService := services.NewTransactionService(userRepo, transRepo)
	auth.SetJWTSecret(cfg.JWTSecret)

	h := handlers.NewHandlers(authService, userService, itemService, transService)

	r := gin.Default()
	r.POST("/api/register", h.Register)
	r.POST("/api/auth", h.Authenticate)
	protected := r.Group("/api").Use(middleware.JWTAuthMiddleware())
	protected.GET("/info", h.GetInfo)
	protected.POST("/sendCoin", h.SendCoin)
	protected.GET("/buy/:item", h.BuyItem)

	cleanup := func() {
		cfg.DB.Exec("TRUNCATE TABLE transactions, inventory, users RESTART IDENTITY CASCADE")
		cfg.Close()
	}

	return r, cfg.DB, cleanup
}

func TestE2EBuyItem(t *testing.T) {
	r, _, cleanup := setupTest(t)
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

func TestE2ESendCoin(t *testing.T) {
	r, _, cleanup := setupTest(t)
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
			Sent     []struct {
				ToUser int `json:"toUser"`
				Amount int `json:"amount"`
			} `json:"sent"`
		} `json:"coinHistory"`
	}
	json.Unmarshal(w.Body.Bytes(), &infoResp)
	if infoResp.Coins != 900 {
		t.Errorf("Ожидалось 900 монет у sender, получено %d", infoResp.Coins)
	}
	if len(infoResp.CoinHistory.Sent) != 1 || infoResp.CoinHistory.Sent[0].Amount != 100 {
		t.Errorf("Ожидалась одна отправленная транзакция на 100, получено %v", infoResp.CoinHistory.Sent)
	}
}
