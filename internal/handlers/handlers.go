package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/itocode21/MerchServiceAvito/internal/config"
	"github.com/itocode21/MerchServiceAvito/internal/database"
	"github.com/itocode21/MerchServiceAvito/internal/services"
)

type Handlers struct {
	config       *config.Config
	authService  *services.AuthService
	userService  *services.UserService
	itemService  *services.ItemService
	transService *services.TransactionService
}

func NewHandlers(config *config.Config, authService *services.AuthService, userService *services.UserService, itemService *services.ItemService, transService *services.TransactionService) *Handlers {
	return &Handlers{
		config:       config,
		authService:  authService,
		userService:  userService,
		itemService:  itemService,
		transService: transService,
	}
}

func (h *Handlers) ResetDB(c *gin.Context) {
	if err := database.ResetDB(h.config.DB); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "База данных очищена"})
}
