package handlers

import "github.com/gin-gonic/gin"

func (h *Handlers) Register(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Неверный запрос"})
		return
	}
	user, err := h.userService.RegisterUser(req.Username, req.Password)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, user)
}

func (h *Handlers) Authenticate(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Неверный запрос"})
		return
	}
	token, err := h.authService.Authenticate(req.Username, req.Password)
	if err != nil {
		c.JSON(401, gin.H{"error": "Неверный логин или пароль"})
		return
	}
	c.JSON(200, gin.H{"token": token})
}
