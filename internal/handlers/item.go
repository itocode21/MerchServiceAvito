package handlers

import (
	"log"

	"github.com/gin-gonic/gin"
)

func (h *Handlers) BuyItem(c *gin.Context) {
	itemName := c.Param("item")
	if itemName == "" {
		log.Printf("BuyItem failed: no item name provided")
		c.JSON(400, gin.H{"error": "Не указано название предмета"})
		return
	}

	username := c.MustGet("username").(string)
	err := h.itemService.BuyItem(username, itemName)
	if err != nil {
		log.Printf("BuyItem failed for user %s, item %s: %v", username, itemName, err)
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	log.Printf("BuyItem succeeded for user %s, item %s", username, itemName)
	c.JSON(200, gin.H{"message": "Предмет успешно куплен"})
}
