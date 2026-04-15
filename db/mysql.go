package db

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

func Connect() (*sql.DB, error) {
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASSWORD")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")

	if dbUser == "" {
		return nil, fmt.Errorf("error: DB_USER is not set")
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", dbUser, dbPass, dbHost, dbPort, dbName)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	todoTableSQL := `
	CREATE TABLE IF NOT EXISTS tasks (
		id INT AUTO_INCREMENT PRIMARY KEY,
		title VARCHAR(255) NOT NULL,
		categorize VARCHAR(255) NOT NULL,
		done BOOLEAN NOT NULL DEFAULT FALSE,
		memo TEXT,
		room_id VARCHAR(255) NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	_, err = db.Exec(todoTableSQL)
	if err != nil {
		return nil, err
	}

	roomHistorySQL := `
	CREATE TABLE IF NOT EXISTS room_history (
		id INT AUTO_INCREMENT PRIMARY KEY,
		room_id VARCHAR(255) NOT NULL UNIQUE,
		last_accessed DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
	);`

	_, err = db.Exec(roomHistorySQL)
	if err != nil {
		return nil, err
	}

	return db, nil
}
