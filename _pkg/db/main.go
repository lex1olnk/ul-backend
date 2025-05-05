package db

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

var Pool *pgxpool.Pool

func Init() error {
	connStr := os.Getenv("POSTGRES_URL")
	if os.Getenv("VERCEL") == "" { // Только для локального окружения
		if err := godotenv.Load(); err != nil {
			log.Fatal("Error loading .env file: ", err)
		}
		connStr = os.Getenv("LOCAL_POSTGRES_URL")
	}

	config, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return fmt.Errorf("error parsing connection string: %w", err)
	}

	Pool, err = pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return fmt.Errorf("unable to create connection pool: %w", err)
	}

	return nil
}

func Close() {
	if Pool != nil {
		Pool.Close()
	}
}
