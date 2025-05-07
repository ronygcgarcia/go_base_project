package routes

import (
    "github.com/gin-gonic/gin"
	"net/http"
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
