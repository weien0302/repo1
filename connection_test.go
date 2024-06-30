package main

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
)

func cleanupDB(t *testing.T, conn *pgx.Conn) {
	_, err := conn.Exec(context.Background(), "DROP TABLE IF EXISTS test_users")
	if err != nil {
		t.Fatalf("Unable to drop test_users table: %v", err)
	}
}

func setupDB(t *testing.T) *pgx.Conn {
	conn, err := pgx.Connect(context.Background(), "postgres://postgres:admin123@localhost:5432/test2")
	if err != nil {
		t.Fatalf("Unable to connect to database: %v", err)
	}

	_, err = conn.Exec(context.Background(), "DROP TABLE IF EXISTS test_users")
	if err != nil {
		t.Fatalf("Unable to drop existing test_users table: %v", err)
	}

	_, err = conn.Exec(context.Background(), "CREATE TABLE test_users (id SERIAL PRIMARY KEY, name TEXT, email TEXT)")
	if err != nil {
		t.Fatalf("Unable to create test_users table: %v", err)
	}

	return conn
}

func TestCRUDOperations(t *testing.T) {
	conn := setupDB(t)
	defer conn.Close(context.Background())
	defer cleanupDB(t, conn)

	// Insert
	_, err := conn.Exec(context.Background(), "INSERT INTO test_users (name, email) VALUES ($1, $2)", "testuser", "testuser@email.com")
	assert.NoError(t, err, "Inserting user should not produce an error")

	// Check insert
	var name, email string
	err = conn.QueryRow(context.Background(), "SELECT name, email FROM test_users WHERE email = $1", "testuser@email.com").Scan(&name, &email)
	assert.NoError(t, err, "Selecting user after insert should not produce an error")
	assert.Equal(t, "testuser", name, "Inserted user name should match")
	assert.Equal(t, "testuser@email.com", email, "Inserted user email should match")

	// Update
	_, err = conn.Exec(context.Background(), "UPDATE test_users SET name = $1 WHERE email = $2", "newName", "testuser@email.com")
	assert.NoError(t, err, "Updating user should not produce an error")

	// Check update
	err = conn.QueryRow(context.Background(), "SELECT name FROM test_users WHERE email = $1", "testuser@email.com").Scan(&name)
	assert.NoError(t, err, "Selecting user after update should not produce an error")
	assert.Equal(t, "newName", name, "Updated user name should match")

	// Delete
	_, err = conn.Exec(context.Background(), "DELETE FROM test_users WHERE email = $1", "testuser@email.com")
	assert.NoError(t, err, "Deleting user should not produce an error")

	// Check delete
	err = conn.QueryRow(context.Background(), "SELECT email FROM test_users WHERE email = $1", "testuser@email.com").Scan(&email)
	assert.Error(t, err, "Selecting user after delete should produce an error")
}
