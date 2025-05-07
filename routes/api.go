package router

import (
    "github.com/gin-gonic/gin"
    "github.com/ronygcgarcia/go_base_project/controllers"
)

func SetupRouter() *gin.Engine {
    r := gin.Default()

    api := r.Group("/api")
    {
        api.GET("/users", controllers.GetAllUsers)
        api.POST("/users", controllers.CreateUser)
    }

    return r
}
