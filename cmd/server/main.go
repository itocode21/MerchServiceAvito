package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/itocode21/MerchServiceAvito/internal/auth"
	"github.com/itocode21/MerchServiceAvito/internal/config"
	"github.com/itocode21/MerchServiceAvito/internal/handlers"
	"github.com/itocode21/MerchServiceAvito/internal/middleware"
	"github.com/itocode21/MerchServiceAvito/internal/repositories"
	"github.com/itocode21/MerchServiceAvito/internal/services"
)

func main() {

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Ошибка загрузки конфигурации: %v", err)
	}
	defer cfg.Close()

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

	// Открытые эндпоинты(не требуют токен)
	r.POST("/api/register", h.Register)
	r.POST("/api/auth", h.Authenticate)

	// Защищённые эндпоинты(требуют токен)
	protected := r.Group("/api").Use(middleware.JWTAuthMiddleware())
	{
		protected.GET("/info", h.GetInfo)
		protected.POST("/sendCoin", h.SendCoin)
		protected.GET("/buy/:item", h.BuyItem)
	}

	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Ошибка запуска сервера: %v", err)
	}
}
