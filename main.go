package main

/* test package golang
split the the code into multiple files
- main.go
- db.go
- handlers.go
- item.go
- routes.go
- utils.go
add additional features from the test
- filter by status and due date
- sort by status, due date, created_at
- partial update
- delete
- error handling
- validation
- unit tests

*/
import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

type Item struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"desc"`
	StatusField string     `json:"status"`
	DueDate     *time.Time `json:"due_date"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

func main() {
	// setup database
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

	// setup handlers
	// /list should not work or should i register both paths?
	http.HandleFunc("/list/", func(w http.ResponseWriter, r *http.Request) {

		handleItems(db, w, r)
	})

	/* simple query
	rows, err := db.Query("SELECT * FROM list LIMIT 1")
	if err != nil {
		log.Fatal("Query failed:", err)
	}
	defer rows.Close()

	if rows.Next() {
		fmt.Println("found a record in the table HOORAY")
	} */

	fmt.Println("Server on 8080, http://localhost:8080/list")
	log.Fatal(http.ListenAndServe(":8080", nil))

}

/*
 table name is list
 database name is todoapi
 host is from cat /etc/resolv.conf if via ip
*/

func setupDB() (*sql.DB, error) {
	connStr := "host=localhost port=5432 user=irkcat password=123 dbname=todoapi sslmode=disable"
	return sql.Open("postgres", connStr)
}

// disable /list route and confine routes to /list/ ?
// CRUD function
func handleItems(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// to-do, doing, completed as status
	// log.Println(r.Method)
	switch r.Method {
	// implement filter by status and due date
	case "GET":

		//single retrieve
		id := strings.TrimPrefix(r.URL.Path, "/list/")
		if id != "" {
			_, err := uuid.Parse(id)
			if err != nil {
				http.Error(w, "Invalid UUID", http.StatusBadRequest)
				return
			}

			var item Item
			err = db.QueryRow("SELECT * FROM list WHERE id = $1", id).Scan(
				&item.ID,
				&item.Title,
				&item.Description,
				&item.StatusField,
				&item.DueDate,
				&item.CreatedAt,
				&item.UpdatedAt,
			)

			if err == sql.ErrNoRows {
				http.Error(w, "Item not found", http.StatusNotFound)
				return
			}

			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			json.NewEncoder(w).Encode(item)
			return
		}

		sortBy := r.URL.Query().Get("sort")
		orderBy := r.URL.Query().Get("order")

		if orderBy == "" {
			orderBy = "ASC"
		}

		query := "SELECT * FROM list"
		switch sortBy {
		case "status":
			query += " ORDER BY status " + orderBy
		case "due_date":
			query += " ORDER BY due_date " + orderBy
		case "created_at":
			query += " ORDER BY created_at " + orderBy
		default:
			query += " ORDER BY created_at	" + orderBy
		}

		rows, err := db.Query(query)
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

		// create
		// can update by post request, maybe reject if there is trailing id present
		// has issues
	case "POST":

		// check path properly
		if r.URL.Path != "/list/" {
			http.Error(w, "Invalid path", http.StatusNotFound)
			//log.Println("171")
			return
		}

		var item Item
		if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if item.Title == "" {
			http.Error(w, "Title required", http.StatusBadRequest)
			return
		}

		if item.Description == "" {
			http.Error(w, "Description Required", http.StatusBadRequest)
			return
		}

		if item.DueDate != nil {
			if item.DueDate.Before(time.Now()) {
				http.Error(w, "Due date cannot be in the past", http.StatusBadRequest)
				return
			}
		}

		query := `
		INSERT INTO list (title, description, status, due_date)
		VALUES ($1, $2, $3, $4)
		RETURNING id, title, description, status, due_date, created_at, updated_at
		`
		err := db.QueryRow(
			query,
			item.Title,
			item.Description,
			item.StatusField,
			item.DueDate,
		).Scan(
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
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(item)

		// update and implement partial update
		// partial update kind of works but will set empty fields to null in json RAW
	case "PUT":
		id := strings.TrimPrefix(r.URL.Path, "/list/")
		if id == "" {
			http.Error(w, "ID is required", http.StatusBadRequest)
			return
		}

		var item Item
		if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if item.DueDate != nil {
			if item.DueDate.Before(time.Now()) {
				http.Error(w, "Due date cannot be in past", http.StatusBadRequest)
				return
			}
		}

		var exists bool
		err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM list WHERE id = $1)", id).Scan(&exists)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if !exists {
			http.Error(w, "Item not found", http.StatusNotFound)
			return
		}

		query := `
        UPDATE list 
        SET title = $1, description = $2, status = $3, due_date = $4, updated_at = NOW()
        WHERE id = $5
        RETURNING id, title, description, status, due_date, created_at, updated_at`

		err = db.QueryRow(
			query,
			item.Title,
			item.Description,
			item.StatusField,
			item.DueDate,
			id,
		).Scan(
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

		json.NewEncoder(w).Encode(item)

	case "DELETE":

		id := strings.TrimPrefix(r.URL.Path, "/list/")
		fmt.Printf("Attempting to delete ID: %s\n", id)

		if id == "" {
			http.Error(w, "UUID is required", http.StatusBadRequest)
			return
		}

		_, err := uuid.Parse(id)
		if err != nil {
			http.Error(w, "invalid UUID", http.StatusBadRequest)
			return
		}

		result, err := db.Exec("DELETE FROM list WHERE id = $1", id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if rowsAffected == 0 {
			http.Error(w, "Todo not found", http.StatusNotFound)
			return
		}

		w.WriteHeader(http.StatusOK)

	}
}
