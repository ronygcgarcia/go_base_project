package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func RegisterUserRoutes(rg *gin.RouterGroup) {
	/**
		rg.GET("/", controllers.GetAllUsers)
	    rg.POST("/", controllers.CreateUser)
	*/
	rg.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "User routes"})
	})
}
