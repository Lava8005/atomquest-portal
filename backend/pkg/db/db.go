/*
Db package initializes and handles the structural lifecycle of our relational data store.
Humanized top-level documentation: Uses connection pool capping to guarantee zero resource leakage.
*/

package db

import (
	"log"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// InitDB builds a highly performant, thread-safe connection pool for PostgreSQL
func InitDB(dataSourceName string) *sqlx.DB {
	// Open the connection using sqlx wrapper for native struct binding capabilities
	db, err := sqlx.Connect("postgres", dataSourceName)
	if err != nil {
		log.Fatalf("Critical Error: Database connection failed: %v", err)
	}

	// Set connection pool parameters tailored for high-concurrency hackathon workloads
	db.SetMaxOpenConns(25)                 // Max active concurrent connections allowed
	db.SetMaxIdleConns(5)                  // Max idle connections kept in pool
	db.SetConnMaxLifetime(5 * time.Minute) // Connection reuse lifespan limit

	// Ensure the datastore is reachable and verified
	if err := db.Ping(); err != nil {
		log.Fatalf("Critical Error: Database ping test failed: %v", err)
	}

	return db
}
