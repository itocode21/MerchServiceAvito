package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/itocode21/MerchServiceAvito/internal/models"
)

func (h *Handlers) GetInfo(c *gin.Context) {
	username := c.MustGet("username").(string)

	cacheKey := "user_info:" + username
	cached, err := h.config.Redis.Get(context.Background(), cacheKey).Result()
	if err == nil {
		c.JSON(http.StatusOK, json.RawMessage(cached))
		return
	}

	info, err := h.userService.GetUserInfo(username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if info == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Пользователь не найден"})
		return
	}

	// Формируем ответ
	response := gin.H{
		"coins":     info.Coins,
		"inventory": info.InventoryJSON,
		"coinHistory": gin.H{
			"received": info.ReceivedJSON,
			"sent":     info.SentJSON,
		},
	}

	responseJSON, err := json.Marshal(response)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to marshal response"})
		return
	}

	// Сохраняем в кэш
	h.config.Redis.Set(context.Background(), cacheKey, responseJSON, 5*time.Minute)
	c.JSON(http.StatusOK, response)
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
