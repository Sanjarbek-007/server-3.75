package postgres

import (
	"balancer/models"
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func InitDB(connStr string) error {
	var err error
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("failed to connect to the database: %v", err)
	}

	if err = DB.Ping(); err != nil {
		return fmt.Errorf("failed to ping the database: %v", err)
	}

	fmt.Println("Database connection established")
	return nil
}

func CloseDB() {
	if DB != nil {
		DB.Close()
	}
}


func RegisterUser(username, email string) (string, error) {
	var userID string
	stmt, err := DB.Prepare("INSERT INTO users(username, email) VALUES($1, $2) RETURNING id")
	if err != nil {
		return "", fmt.Errorf("prepare statement error: %v", err)
	}
	defer stmt.Close()

	err = stmt.QueryRow(username, email).Scan(&userID)
	if err != nil {
		return "", fmt.Errorf("query row error: %v", err)
	}

	return userID, nil
}

func ListUsers() ([]models.User, error) {
	rows, err := DB.Query(`
		SELECT id, username, email FROM users
		UNION ALL
		SELECT id, username, email FROM users_server2
	`)
	if err != nil {
		return nil, fmt.Errorf("query error: %v", err)
	}
	defer rows.Close()

	users := []models.User{}
	for rows.Next() {
		var user models.User
		if err := rows.Scan(&user.ID, &user.Username, &user.Email); err != nil {
			return nil, fmt.Errorf("scan error: %v", err)
		}
		users = append(users, user)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %v", err)
	}
	return users, nil
}


func UpdateUser(id, username, email string) (bool, error) {
	tx, err := DB.Begin()
	if err != nil {
		return false, fmt.Errorf("transaction begin error: %v", err)
	}

	stmt1, err := tx.Prepare("UPDATE users SET username=$1, email=$2 WHERE id=$3")
	if err != nil {
		tx.Rollback()
		return false, fmt.Errorf("prepare statement error for server 2: %v", err)
	}
	defer stmt1.Close()

	res1, err := stmt1.Exec(username, email, id)
	if err != nil {
		tx.Rollback()
		return false, fmt.Errorf("execute statement error for server 2: %v", err)
	}

	stmt2, err := tx.Prepare("UPDATE users_server2 SET username=$1, email=$2 WHERE id=$3")
	if err != nil {
		tx.Rollback()
		return false, fmt.Errorf("prepare statement error for server 1: %v", err)
	}
	defer stmt2.Close()

	res2, err := stmt2.Exec(username, email, id)
	if err != nil {
		tx.Rollback()
		return false, fmt.Errorf("execute statement error for server 1: %v", err)
	}

	rowsAffected1, err := res1.RowsAffected()
	if err != nil || rowsAffected1 == 0 {
		tx.Rollback()
		return false, fmt.Errorf("no rows affected for server 2")
	}

	rowsAffected2, err := res2.RowsAffected()
	if err != nil || rowsAffected2 == 0 {
		tx.Rollback()
		return false, fmt.Errorf("no rows affected for server 1")
	}

	if err := tx.Commit(); err != nil {
		return false, fmt.Errorf("transaction commit error: %v", err)
	}

	return true, nil
}

func DeleteUser(id string) (bool, error) {
	tx, err := DB.Begin()
	if err != nil {
		return false, fmt.Errorf("transaction begin error: %v", err)
	}

	stmt1, err := tx.Prepare("DELETE FROM users WHERE id=$1")
	if err != nil {
		tx.Rollback()
		return false, fmt.Errorf("prepare statement error for server 1: %v", err)
	}
	defer stmt1.Close()

	res1, err := stmt1.Exec(id)
	if err != nil {
		tx.Rollback()
		return false, fmt.Errorf("execute statement error for server 1: %v", err)
	}

	stmt2, err := tx.Prepare("DELETE FROM users_server2 WHERE id=$1")
	if err != nil {
		tx.Rollback()
		return false, fmt.Errorf("prepare statement error for server 1: %v", err)
	}
	defer stmt2.Close()

	res2, err := stmt2.Exec(id)
	if err != nil {
		tx.Rollback()
		return false, fmt.Errorf("execute statement error for server 1: %v", err)
	}

	rowsAffected1, err := res1.RowsAffected()
	if err != nil || rowsAffected1 == 0 {
		tx.Rollback()
		return false, fmt.Errorf("no rows affected for server 2")
	}

	rowsAffected2, err := res2.RowsAffected()
	if err != nil || rowsAffected2 == 0 {
		tx.Rollback()
		return false, fmt.Errorf("no rows affected for server 1")
	}

	if err := tx.Commit(); err != nil {
		return false, fmt.Errorf("transaction commit error: %v", err)
	}

	return true, nil
}