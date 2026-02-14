package store

import "github.com/jmoiron/sqlx"

// Store wraps a sqlx database connection and provides data access methods.
type Store struct {
	db *sqlx.DB
}

// New creates a new Store.
func New(db *sqlx.DB) *Store {
	return &Store{db: db}
}
