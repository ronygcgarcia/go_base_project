package routes

import (
	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	api := r.Group("/api")
	{
		// RegisterAuthRoutes(api.Group("/auth"))
		RegisterUserRoutes(api.Group("/users"))
	}

	return r
}
