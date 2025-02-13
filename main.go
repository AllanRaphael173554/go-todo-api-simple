package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	_ "github.com/lib/pq"
)

type Item struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"desc"`
	StatusField string    `json:"status"`
	DueDate     time.Time `json:"due_date"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func main() {
	db, err := setupDB()
	if err != nil {
		log.Fatal("Failled to establish connection: ", err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Fatal("Failed to ping db", err)
	}
	fmt.Println("Connection success")

}

func setupDB() (*sql.DB, error) {
	// table name is list
	connStr := "postgresql://username:password@localhost/dbname?sslmode=disable"
	return sql.Open("postgres", connStr)
}

func handleItems(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case "GET":

	case "POST":
		// to-do, doing, completed as status
	case "PUT":

	case "DELETE":

	}
}
