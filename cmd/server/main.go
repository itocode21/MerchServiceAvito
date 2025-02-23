package main

import (
	"log"

	"github.com/joho/godotenv"

	"github.com/gin-gonic/gin"
	"github.com/itocode21/MerchServiceAvito/internal/database"
	"github.com/itocode21/MerchServiceAvito/internal/repositories"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Ошибка загрузки .env файла: %v", err)
	}

	db, err := database.NewDB()
	if err != nil {
		log.Fatalf("Ошибка подключения к базе данных: %v", err)
	}
	defer db.Close()

	itemRepo := repositories.NewItemRepository(db)

	r := gin.Default()

	r.GET("api/info", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "Yo, Avito!"})
	})
	r.GET("/api/items", func(c *gin.Context) {
		items, err := itemRepo.GetAllItems()
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, items)
	})

	r.Run(":8080")
}
