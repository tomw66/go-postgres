package database

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "123456789"
	dbname   = "postgres"
)

func Script() {
	db := InitialiseTable()

	defer db.Close()

	// Create a new record
	id, err := CreateRecord(db, "John", 30)
	if err != nil {
		log.Fatal("Error creating record: ", err)
	}
	fmt.Println("Created record with ID:", id)

	// Read records
	records, err := ReadRecords(db)
	if err != nil {
		log.Fatal("Error reading records: ", err)
	}
	fmt.Println("All records:")
	for _, r := range records {
		fmt.Println("ID:", r.ID, "Name:", r.Name, "Age:", r.Age)
	}

	// Update a record
	err = UpdateRecord(db, id, "Janey", 35)
	if err != nil {
		log.Fatal("Error updating record: ", err)
	}
	fmt.Println("Updated record with ID:", id)

	// Delete a record
	err = DeleteRecord(db, id)
	if err != nil {
		log.Fatal("Error deleting record: ", err)
	}
	fmt.Println("Deleted record with ID:", id)
}

// Record represents a database record
type Record struct {
	ID   int
	Name string
	Age  int
}

// createEmptyTable creates the records table if it doesn't exist
func createEmptyTable(db *sql.DB) error {
    query := `
    CREATE TABLE IF NOT EXISTS records (
        id SERIAL PRIMARY KEY,
        name VARCHAR(100),
        age INT,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
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
// TODO delete in favour of InsertRecord?
func CreateRecord(db *sql.DB, name string, age int) (int, error) {
	var id int
	err := db.QueryRow("INSERT INTO records(name, age) VALUES($1, $2) RETURNING id", name, age).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

// ReadRecords retrieves all records from the database
func ReadRecords(db *sql.DB) ([]Record, error) {
	rows, err := db.Query("SELECT id, name, age FROM records")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []Record
	for rows.Next() {
		var r Record
		err := rows.Scan(&r.ID, &r.Name, &r.Age)
		if err != nil {
			return nil, err
		}
		records = append(records, r)
	}
	return records, nil
}

// ClearRecords removes all records from the database
func ClearRecords(db *sql.DB) error {
	_, err := db.Exec("DELETE FROM records")
	return err
}

// InsertRecord inserts a single record into the database
func InsertRecord(db *sql.DB, record Record) error {
	_, err := db.Exec("INSERT INTO records (id, name, age) VALUES ($1, $2, $3)", record.ID, record.Name, record.Age)
	return err
}

// UpdateRecord updates an existing record in the database
func UpdateRecord(db *sql.DB, id int, name string, age int) error {
	_, err := db.Exec("UPDATE records SET name=$1, age=$2 WHERE id=$3", name, age, id)

	return err
}

func DeleteRecord(db *sql.DB, id int) error {
	_, err := db.Exec("DELETE FROM records WHERE id=$1", id)

	return err
}