package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v4/pgxpool"
)

// Storage - data storage connected to the PostgreSQL database via a connection pool.
type PostgresStorage struct {
	db *pgxpool.Pool
}

// New creates a new connection to the PostgreSQL database.
func New(content string) (*PostgresStorage, error) {
	db, err := pgxpool.Connect(context.Background(), content)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	err = db.Ping(context.Background())
	if err != nil {
		return nil, fmt.Errorf("cannot ping PostgreSQL: %v", err)
	}

	ps := PostgresStorage{
		db: db,
	}
	return &ps, nil
}

// Close closes the pool of connections to the PostgreSQL database.
func (ps *PostgresStorage) Close() {
	ps.db.Close()
}
