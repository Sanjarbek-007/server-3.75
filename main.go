package main

import (
	"fmt"
	"log"
	"net/http"

	"balancer/models"
	"balancer/storage/postgres"

	"github.com/gin-gonic/gin"
)

func main() {
	err := postgres.InitDB("host=postgres1 port=5432 user=postgres password=1111 dbname=server1_db sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer postgres.CloseDB()

	r := gin.Default()

	r.POST("/user/register", registerUser)
	r.GET("/user/list", listUsers)
	r.PUT("/user/update/:id", updateUser)
	r.DELETE("/user/delete/:id", deleteUser)
	r.GET("/health", health)

	fmt.Println("Server is running on :8081")
	if err := r.Run(":8081"); err != nil {
		log.Fatal(err)
	}
}

func registerUser(c *gin.Context) {
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, err := postgres.RegisterUser(user.Username, user.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	user.ID = userID
	c.JSON(http.StatusCreated, user)
}

func listUsers(c *gin.Context) {
	users, err := postgres.ListUsers()
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

	updated, err := postgres.UpdateUser(id, user.Username, user.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if updated {
		c.JSON(http.StatusOK, gin.H{"message": "User updated successfully"})
	} else {
		c.JSON(http.StatusNotFound, gin.H{"message": "User not found"})
	}
}

func deleteUser(c *gin.Context) {
	id := c.Param("id")

	updated, err := postgres.DeleteUser(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if updated {
		c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
	} else {
		c.JSON(http.StatusNotFound, gin.H{"message": "User not found"})
	}
}

func health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "OK"})
}
