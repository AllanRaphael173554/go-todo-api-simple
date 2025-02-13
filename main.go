package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	_ "github.com/lib/pq"
)

type Item struct {
	ID          string    `json:"id"`
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

	// ping db
	err = db.Ping()
	if err != nil {
		log.Fatal("Failed to ping db", err)
	}
	fmt.Println("Connection success")

	// setup handler
	http.HandleFunc("/list", func(w http.ResponseWriter, r *http.Request) {

		handleItems(db, w, r)

	})

	// simple query
	rows, err := db.Query("SELECT * FROM list LIMIT 1")
	if err != nil {
		log.Fatal("Query failed:", err)
	}
	defer rows.Close()

	if rows.Next() {
		fmt.Println("found a record in the table")
	}

	fmt.Println("Server on 5432")
	log.Fatal(http.ListenAndServe(":5432", nil))

}

func setupDB() (*sql.DB, error) {
	/*
		table name is list
		database name is todoapi
		host is from cat /etc/resolv.conf
	*/
	connStr := "host=localhost port=5432 user=postgres password=123 dbname=todoapi sslmode=disable"
	return sql.Open("postgres", connStr)
}

// CRUD function
func handleItems(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// to-do, doing, completed as status
	switch r.Method {
	case "GET":
		rows, err := db.Query("SELECT * FROM list")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var items []Item
		for rows.Next() {
			var item Item
			err := rows.Scan(
				&item.ID,
				&item.Title,
				&item.Description,
				&item.StatusField,
				&item.DueDate,
				&item.CreatedAt,
				&item.UpdatedAt,
			)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			items = append(items, item)
		}
		json.NewEncoder(w).Encode(items)

	case "POST":

	case "PUT":

	case "DELETE":

	}
}
