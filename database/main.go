package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "123456789"
	dbname   = "postgres"
)

// Record represents a database record
type Record struct {
	ID       int
	Priority int
	Task     string
	Due      time.Time
}

// createEmptyTable creates the table if it doesn't exist
func createEmptyTable(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS todo (
		id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
		priority INT,
		task VARCHAR(100),
		due TIMESTAMP
	)`
	_, err := db.Exec(query)
	return err
}

func InitialiseTable() *sql.DB {
	// Establish a connection to the PostgreSQL database
	db, err := sql.Open("postgres", fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname))
	if err != nil {
		log.Fatal("Error connecting to the database: ", err)
	}

	// Create the table if it doesn't exist
	err = createEmptyTable(db)
	if err != nil {
		log.Fatal("Error creating table: ", err)
	}

	return db
}

// CreateRecord creates a new record in the database
func CreateRecord(db *sql.DB, priority int, task string, due time.Time) (int, error) {
	var id int
	err := db.QueryRow("INSERT INTO todo(priority, task, due) VALUES($1, $2, $3) RETURNING id", priority, task, due).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

// ReadRecords retrieves all records from the database
func ReadRecords(db *sql.DB) ([]Record, error) {
	rows, err := db.Query("SELECT id, priority, task, due FROM todo")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []Record
	for rows.Next() {
		var r Record
		err := rows.Scan(&r.ID, &r.Priority, &r.Task, &r.Due)
		if err != nil {
			return nil, err
		}
		records = append(records, r)
	}
	return records, nil
}

// ClearRecords removes all records from the database
func ClearRecords(db *sql.DB) error {
	_, err := db.Exec("DELETE FROM todo")
	return err
}

// UpdateRecord updates an existing record in the database
func UpdateRecord(db *sql.DB, id int, priority int, task string, due time.Time) error {
	_, err := db.Exec("UPDATE todo SET priority=$2, task=$3, due=$4 WHERE id=$1", id, priority, task, due)

	return err
}

func DeleteRecord(db *sql.DB, id int) error {
	_, err := db.Exec("DELETE FROM todo WHERE id=$1", id)

	return err
}