package handlers

import "github.com/gin-gonic/gin"

func (h *Handlers) SendCoin(c *gin.Context) {
	var req struct {
		ToUser string `json:"toUser"`
		Amount int    `json:"amount"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Неверный запрос"})
		return
	}
	fromUser := c.MustGet("username").(string)
	err := h.transService.SendCoins(fromUser, req.ToUser, req.Amount)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "Монеты успешно отправлены"})
}
