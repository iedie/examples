// sql provides an SQL implementation of our Storer interface (specifically PostGreSQL).
package sql

import (
	"database/sql"

	// Load postgres driver
	_ "github.com/lib/pq"
)

// DB implements Storer using a PostGreSQL database.
type DB struct {
	storage *sql.DB // Here we simply refer to it as "storage" to avoid common naming conflicts
}

// NewSQLDB creates a new database connection for use.
func NewSQLDB(url string) (*DB, error) {
	// Connect to database with supplied URL
	db, err := sql.Open("postgres", url)
	if err != nil {
		return nil, err
	}
	// Ensure connection is usable
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}
	// Usable connection, return it for use
	return &DB{storage: db}, nil
}

// Session and User methods can be found in their respective files (session.go, user.go)
