package db

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/lib/pq"
)

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository() (*PostgresRepository, error) {

	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	yourdb := os.Getenv("DB_NAME")
	username := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")

	//"postgres://username:password@localhost:5432/yourdb"
	connString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", username, password, host, port, yourdb)

	pool, err := pgxpool.New(context.Background(), connString)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &PostgresRepository{pool: pool}, nil
}

func (r *PostgresRepository) Close() {
	if r.pool != nil {
		r.pool.Close()
	}
}

// Execute executes a query that doesn't return rows
func (r *PostgresRepository) Execute(ctx context.Context, query string, args ...interface{}) (pgconn.CommandTag, error) {
	return r.pool.Exec(ctx, query, args...)
}

// Query executes a query that returns rows
func (r *PostgresRepository) Query(ctx context.Context, query string, args ...interface{}) (pgx.Rows, error) {
	return r.pool.Query(ctx, query, args...)
}

// QueryRow executes a query that returns at most one row
func (r *PostgresRepository) QueryRow(ctx context.Context, query string, args ...interface{}) pgx.Row {
	return r.pool.QueryRow(ctx, query, args...)
}

// Transaction executes function within a transaction
func (r *PostgresRepository) Transaction(ctx context.Context, fn func(pgx.Tx) error) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback(ctx)
			panic(p)
		}
	}()

	if err := fn(tx); err != nil {
		tx.Rollback(ctx)
		return err
	}

	return tx.Commit(ctx)
}
