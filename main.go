package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"balancer/storage/postgres"

	"github.com/gin-gonic/gin"
)

type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

func main() {
	if err := postgres.InitDB(); err != nil {
		log.Fatal(err)
	}

	r := gin.Default()

	r.POST("/user/register", registerUser)
	r.GET("/user/list", listUsers)
	r.PUT("/user/update/:id", updateUser)
	r.GET("/health", health)

	fmt.Println("Server is running on :8081")
	if err := r.Run("auth:8081"); err != nil {
		log.Fatal(err)
	}
}

func registerUser(c *gin.Context) {
	var user User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if _, err := postgres.RegisterUser(user.Username, user.Email); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		log.Printf("RegisterUser error: %v", err)
		return
	}

	c.JSON(http.StatusCreated, user)
}

func listUsers(c *gin.Context) {
	users, err := postgres.ListUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		log.Printf("ListUsers error: %v", err)
		return
	}

	c.JSON(http.StatusOK, users)
}

func updateUser(c *gin.Context) {
	idd := c.Param("id")
	id, _ := strconv.Atoi(idd)
	
	var user User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := postgres.UpdateUser(id, user.Username, user.Email); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		log.Printf("UpdateUser error: %v", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User updated successfully"})
}

func health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"Server": "OK"})
}
