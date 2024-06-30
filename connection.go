package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5"
)

type User struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

func main() {
	conn, err := setupDatabase()
	if err != nil {
		handleError("Setup database failed", err)
	}
	defer conn.Close(context.Background())

	r := mux.NewRouter()

	// Setup routes
	r.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			var user User
			if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			if err := insertUser(conn, user); err != nil {
				handleError("Insert user failed", err)
			}
			w.WriteHeader(http.StatusCreated)
		case http.MethodGet:
			if err := listUsers(conn, w); err != nil {
				handleError("List users failed", err)
			}
		}
	}).Methods(http.MethodPost, http.MethodGet)

	r.HandleFunc("/users/{email}", func(w http.ResponseWriter, r *http.Request) {
		email := mux.Vars(r)["email"]
		switch r.Method {
		case http.MethodPut:
			var user User
			if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			if err := updateUser(conn, user.Name, email); err != nil {
				handleError("Update user failed", err)
			}
			w.WriteHeader(http.StatusOK)
		case http.MethodDelete:
			if err := deleteUser(conn, email); err != nil {
				handleError("Delete user failed", err)
			}
			w.WriteHeader(http.StatusOK)
		}
	}).Methods(http.MethodPut, http.MethodDelete)

	// Start server
	fmt.Println("Server is running at http://localhost:8080")
	http.ListenAndServe(":8080", r)
}

func setupDatabase() (*pgx.Conn, error) {
	return pgx.Connect(context.Background(), "postgres://postgres:admin123@localhost:5432/test2")
}

// func createTable(conn *pgx.Conn) error {
// 	_, err := conn.Exec(context.Background(), "CREATE TABLE IF NOT EXISTS users (name TEXT, email TEXT)")
// 	return err
// }

func insertUser(conn *pgx.Conn, user User) error {
	_, err := conn.Exec(context.Background(), "INSERT INTO users (name, email) VALUES ($1, $2)", user.Name, user.Email)
	return err
}

func updateUser(conn *pgx.Conn, newName, email string) error {
	_, err := conn.Exec(context.Background(), "UPDATE users SET name = $1 WHERE email = $2", newName, email)
	return err
}

func deleteUser(conn *pgx.Conn, email string) error {
	_, err := conn.Exec(context.Background(), "DELETE FROM users WHERE email = $1", email)
	return err
}

func handleError(msg string, err error) {
	fmt.Fprintf(os.Stderr, "%s: %v\n", msg, err)
	os.Exit(1)
}

func listUsers(conn *pgx.Conn, w http.ResponseWriter) error {
	rows, err := conn.Query(context.Background(), "SELECT name, email FROM users")
	if err != nil {
		return err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		if err := rows.Scan(&user.Name, &user.Email); err != nil {
			return err
		}
		users = append(users, user)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)

	return rows.Err()
}
