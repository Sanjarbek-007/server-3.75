package postgres

import (
	"balancer/models"
	"database/sql"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func Connect(dsn string) error {
	var err error
	DB, err = sql.Open("postgres", dsn)
	if err != nil {
		return err
	}
	return DB.Ping()
}

func InitDB() error {
	dsn := "host=localhost port=5432 user=postgres password=1111 dbname=server1_db sslmode=disable"
	return Connect(dsn)
}

func RegisterUser(username, email string) (int, error) {
	var userID int
	stmt, err := DB.Prepare("INSERT INTO users(username, email) VALUES($1, $2) RETURNING id")
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	err = stmt.QueryRow(username, email).Scan(&userID)
	if err != nil {
		return 0, err
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

func UpdateUser(id int, username, email string) error {
	tx, err := DB.Begin()
	if err != nil {
		return err
	}

	stmt1, err := tx.Prepare("UPDATE users SET username=$1, email=$2 WHERE id=$3")
	if err != nil {
		tx.Rollback()
		return err
	}
	defer stmt1.Close()

	res1, err := stmt1.Exec(username, email, id)
	if err != nil {
		tx.Rollback()
		return err
	}

	rowsAffected1, err := res1.RowsAffected()
	if err != nil {
		tx.Rollback()
		return err
	}

	if rowsAffected1 > 0 {
		return tx.Commit()
	}

	stmt2, err := tx.Prepare("UPDATE users_server2 SET username=$1, email=$2 WHERE id=$3")
	if err != nil {
		tx.Rollback()
		return err
	}
	defer stmt2.Close()

	res2, err := stmt2.Exec(username, email, id)
	if err != nil {
		tx.Rollback()
		return err
	}

	rowsAffected2, err := res2.RowsAffected()
	if err != nil {
		tx.Rollback()
		return err
	}

	if rowsAffected2 == 0 {
		tx.Rollback()
		return nil
	}

	return tx.Commit()
}
