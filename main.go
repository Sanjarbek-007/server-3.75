package main

import (
	"log"
	"net/http"

	"balancer/models"
	"balancer/storage/postgres"

	"github.com/gin-gonic/gin"
)

var (
	db1 *postgres.DB
)

func main() {
	var err error

	db1, err = postgres.InitDB("postgres1", "5432", "postgres", "1111", "server1_db")
	if err != nil {
		log.Fatal(err)
	}
	defer db1.DB.Close()

	r := gin.Default()

	r.POST("/user/register", registerUser)
	r.GET("/user/list", listUsers)
	r.PUT("/user/update/:id", updateUser)
	r.GET("/health", health)

	if err := r.Run("auth:8081"); err != nil {
		log.Fatal(err)
	}
}

func registerUser(c *gin.Context) {
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := db1.RegisterUser(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, user)
}

func listUsers(c *gin.Context) {
	users, err := db1.ListUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, users)
}

func updateUser(c *gin.Context) {
	id := c.Param("id")
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	rowsAffected, err := db1.UpdateUser(id, &user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"message": "User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User updated successfully"})
}

func health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"Server": "OK"})
}
