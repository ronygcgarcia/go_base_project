package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ronygcgarcia/go_base_project/config"
	"github.com/ronygcgarcia/go_base_project/models"
)

func GetAllUsers(c *gin.Context) {
	var users []models.User
	config.DB.Find(&users)
	c.JSON(http.StatusOK, users)
}

func CreateUser(c *gin.Context) {
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	config.DB.Create(&user)
	c.JSON(http.StatusCreated, user)
}
