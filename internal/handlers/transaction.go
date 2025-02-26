package handlers

import (
	"log"

	"github.com/gin-gonic/gin"
)

func (h *Handlers) SendCoin(c *gin.Context) {
	var req struct {
		ToUser string `json:"toUser"`
		Amount int    `json:"amount"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("SendCoin failed: invalid request: %v", err)
		c.JSON(400, gin.H{"error": "Неверный запрос"})
		return
	}
	fromUser := c.MustGet("username").(string)
	err := h.transService.SendCoins(fromUser, req.ToUser, req.Amount)
	if err != nil {
		log.Printf("SendCoin failed for user %s to %s, amount %d: %v", fromUser, req.ToUser, req.Amount, err)
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	log.Printf("SendCoin succeeded for user %s to %s, amount %d", fromUser, req.ToUser, req.Amount)
	c.JSON(200, gin.H{"message": "Монеты успешно отправлены"})
}
