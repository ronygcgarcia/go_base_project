package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ronygcgarcia/go_base_project/middlewares"
)

func RegisterUserRoutes(rg *gin.RouterGroup) {
	rg.GET("/", middlewares.AuthRequired("user"), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "User routes"})
	})
}
