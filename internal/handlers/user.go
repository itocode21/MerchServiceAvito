package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/itocode21/MerchServiceAvito/internal/models"
)

func (h *Handlers) GetInfo(c *gin.Context) {
	username := c.MustGet("username").(string)

	user, err := h.userService.GetUserByUsername(username)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	if user == nil {
		c.JSON(404, gin.H{"error": "Пользователь не найден"})
		return
	}

	inventory, err := h.itemService.GetUserInventory(user.ID)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	transactions, err := h.transService.GetUserTransactions(user.ID)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	response := gin.H{
		"coins":     user.Coins,
		"inventory": inventory,
		"coinHistory": gin.H{
			"received": filterTransactions(transactions, user.ID, true),
			"sent":     filterTransactions(transactions, user.ID, false),
		},
	}

	c.JSON(200, response)
}

func filterTransactions(transactions []models.Transaction, userID int, received bool) []gin.H {
	var result []gin.H
	for _, t := range transactions {
		if received && t.ToUserID == userID {
			result = append(result, gin.H{
				"fromUser": t.FromUserID,
				"amount":   t.Amount,
			})
		} else if !received && t.FromUserID == userID {
			result = append(result, gin.H{
				"toUser": t.ToUserID,
				"amount": t.Amount,
			})
		}
	}
	return result
}
