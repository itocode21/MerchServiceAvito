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

	userRepo := repositories.NewUserRepository(cfg)
	itemRepo := repositories.NewItemRepository(cfg.DB)
	transRepo := repositories.NewTransactionRepository(cfg.DB)

	authService := services.NewAuthService(userRepo)
	userService := services.NewUserService(userRepo)
	itemService := services.NewItemService(itemRepo, userRepo)
	transService := services.NewTransactionService(userRepo, transRepo)

	auth.SetJWTSecret(cfg.JWTSecret)

	h := handlers.NewHandlers(cfg, authService, userService, itemService, transService)

	r := gin.Default()
	r.GET("/api/reset", h.ResetDB) // Добавленный эндпоинт для сброса
	r.POST("/api/register", h.Register)
	r.POST("/api/auth", h.Authenticate)
	protected := r.Group("/api").Use(middleware.JWTAuthMiddleware())
	protected.GET("/info", h.GetInfo)
	protected.POST("/sendCoin", h.SendCoin)
	protected.GET("/buy/:item", h.BuyItem)

	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Ошибка запуска сервера: %v", err)
	}
}
