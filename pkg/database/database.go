package database

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Database struct {
	Pool *pgxpool.Pool
}

type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
}

func NewDatabase(config Config) (*Database, error) {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		config.User, config.Password, config.Host, config.Port, config.DBName)

	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		log.Printf("Failed to connect to database: %v", err)
		return nil, err
	}

	// Test the connection
	if err := pool.Ping(context.Background()); err != nil {
		log.Printf("Failed to ping database: %v", err)
		return nil, err
	}

	return &Database{Pool: pool}, nil
}

func (db *Database) Close() {
	if db.Pool != nil {
		db.Pool.Close()
	}
}
