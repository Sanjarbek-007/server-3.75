package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

var db *sql.DB

func main() {
	var err error
	db, err = sql.Open("postgres", "host=postgres1 port=5432 user=postgres password=1111 dbname=server1_db sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err = db.Ping(); err != nil {
		log.Fatal(err)
	}

	r := gin.Default()

	r.POST("/user/register", registerUser)
	r.GET("/user/list", listUsers)
	r.PUT("/user/update/:id", updateUser)  // Add update route
	r.GET("/health", health)

	fmt.Println("Server 1 is running on :8081")
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

	stmt, err := db.Prepare("INSERT INTO users(username, email) VALUES($1, $2) RETURNING id")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		log.Printf("Prepare error: %v", err)
		return
	}
	defer stmt.Close()

	err = stmt.QueryRow(user.Username, user.Email).Scan(&user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		log.Printf("QueryRow error: %v", err)
		return
	}

	c.JSON(http.StatusCreated, user)
}

func listUsers(c *gin.Context) {
	rows, err := db.Query(`
		SELECT id, username, email FROM users
		UNION ALL
		SELECT id, username, email FROM users_server2
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		log.Printf("Query error: %v", err)
		return
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		if err := rows.Scan(&user.ID, &user.Username, &user.Email); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			log.Printf("Scan error: %v", err)
			return
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		log.Printf("Rows error: %v", err)
		return
	}

	c.JSON(http.StatusOK, users)
}

func updateUser(c *gin.Context) {
	id := c.Param("id")
	var user User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tx, err := db.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		log.Printf("Transaction Begin error: %v", err)
		return
	}

	stmt1, err := tx.Prepare("UPDATE users SET username=$1, email=$2 WHERE id=$3")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		log.Printf("Prepare error for server 1: %v", err)
		tx.Rollback()
		return
	}
	defer stmt1.Close()

	res1, err := stmt1.Exec(user.Username, user.Email, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		log.Printf("Exec error for server 1: %v", err)
		tx.Rollback()
		return
	}

	rowsAffected1, err := res1.RowsAffected()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		log.Printf("RowsAffected error for server 1: %v", err)
		tx.Rollback()
		return
	}

	if rowsAffected1 > 0 {
		if err := tx.Commit(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			log.Printf("Commit error: %v", err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "User updated successfully in server 1"})
		return
	}

	stmt2, err := tx.Prepare("UPDATE users_server2 SET username=$1, email=$2 WHERE id=$3")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		log.Printf("Prepare error for server 2: %v", err)
		tx.Rollback()
		return
	}
	defer stmt2.Close()

	res2, err := stmt2.Exec(user.Username, user.Email, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		log.Printf("Exec error for server 2: %v", err)
		tx.Rollback()
		return
	}

	rowsAffected2, err := res2.RowsAffected()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		log.Printf("RowsAffected error for server 2: %v", err)
		tx.Rollback()
		return
	}

	if rowsAffected2 == 0 {
		c.JSON(http.StatusNotFound, gin.H{"message": "User not found in either database"})
		tx.Rollback()
		return
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		log.Printf("Commit error: %v", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User updated successfully in server 2"})
}



func health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"Server":"OK"})
}
