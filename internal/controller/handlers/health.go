package handlers

import (
	"github.com/gin-gonic/gin"
)

// @BasePath /health
// PingExample godoc
// @Summary Application health
// @Schemes
// @Description Check the application health
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} object{status=string} "health response"
// @Failure 404 {object} object{error=string} "not found error response"
// @Failure 500 {object} object{error=string} "generic error response"
// @Router /health [get]
func HandleHealth(c *gin.Context) {
	c.JSON(200, gin.H{
		"status": "UP",
	})
}
