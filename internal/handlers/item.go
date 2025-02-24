package handlers

import (
	"github.com/gin-gonic/gin"
)

func (h *Handlers) BuyItem(c *gin.Context) {
	itemName := c.Param("item")
	if itemName == "" {
		c.JSON(400, gin.H{"error": "Не указано название предмета"})
		return
	}

	username := c.MustGet("username").(string)
	err := h.itemService.BuyItem(username, itemName)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "Предмет успешно куплен"})
}
