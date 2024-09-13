package postgres

import (
	"balancer/models"
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

type DB struct {
	DB *sql.DB
}

func InitDB(host, port, user, password, dbname string) (*DB, error) {
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &DB{DB: db}, nil
}


func (db *DB) RegisterUser(user *models.User) error {
	stmt, err := db.DB.Prepare("INSERT INTO users(username, email) VALUES($1, $2) RETURNING id")
	if err != nil {
		return err
	}
	defer stmt.Close()

	err = stmt.QueryRow(user.Username, user.Email).Scan(&user.ID)
	if err != nil {
		return err
	}

	return nil
}


func (db *DB) ListUsers() ([]models.User, error) {
	rows, err := db.DB.Query("SELECT id, username, email FROM users")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		if err := rows.Scan(&user.ID, &user.Username, &user.Email); err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}


func (db *DB) UpdateUser(id string, user *models.User) (int64, error) {
	stmt, err := db.DB.Prepare("UPDATE users SET username=$1, email=$2 WHERE id=$3")
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	res, err := stmt.Exec(user.Username, user.Email, id)
	if err != nil {
		return 0, err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}

	return rowsAffected, nil
}
