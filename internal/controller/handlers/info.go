package handlers

import "github.com/gin-gonic/gin"

var ProjectName string
var Version string

// @BasePath /info
// PingExample godoc
// @Summary Application info
// @Schemes
// @Description Check the application info(name and version)
// @Tags info
// @Accept json
// @Produce json
// @Success 200 {object} object{name=string,version=string} "info response"
// @Failure 500 {object} object{error=string} "generic error response"
// @Router /info [get]
func HandleInfo(c *gin.Context) {
	c.JSON(200, gin.H{
		"name":    ProjectName,
		"version": Version,
	})
}
