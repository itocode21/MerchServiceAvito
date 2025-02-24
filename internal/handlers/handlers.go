package handlers

import (
	"github.com/itocode21/MerchServiceAvito/internal/services"
)

type Handlers struct {
	authService  *services.AuthService
	userService  *services.UserService
	itemService  *services.ItemService
	transService *services.TransactionService
}

func NewHandlers(auth *services.AuthService, user *services.UserService, item *services.ItemService, trans *services.TransactionService) *Handlers {
	return &Handlers{authService: auth, userService: user, itemService: item, transService: trans}
}
