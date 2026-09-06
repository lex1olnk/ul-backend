package db

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

var Pool *pgxpool.Pool

var (
	once    sync.Once
	initErr error
)

// Init создаёт пул соединений один раз на процесс.
//
// Serverless-инстанс переиспользуется между запросами и обслуживает их
// конкурентно, поэтому пул нельзя создавать и закрывать на каждый запрос:
// Close() из одного запроса обрывал бы соединения остальных.
// Повторные вызовы возвращают уже готовый пул (или ошибку первой попытки).
func Init() error {
	once.Do(func() {
		// .env читаем до os.Getenv, иначе локально переменные ещё не выставлены
		if os.Getenv("VERCEL") == "" { // Только для локального окружения
			if err := godotenv.Load(); err != nil {
				// Отсутствие .env не фатально: переменные могут быть в окружении
				fmt.Println("warning: .env file not loaded:", err)
			}
		}

		connStr := os.Getenv("POSTGRES_URL")
		if connStr == "" {
			initErr = fmt.Errorf("POSTGRES_URL is not set")
			return
		}

		config, err := pgxpool.ParseConfig(connStr)
		if err != nil {
			initErr = fmt.Errorf("error parsing connection string: %w", err)
			return
		}

		// Ограничения под serverless: много коротких инстансов на одну БД
		config.MaxConns = 5
		config.MinConns = 0
		config.MaxConnLifetime = 5 * time.Minute
		config.MaxConnIdleTime = 1 * time.Minute

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		pool, err := pgxpool.NewWithConfig(ctx, config)
		if err != nil {
			initErr = fmt.Errorf("unable to create connection pool: %w", err)
			return
		}

		Pool = pool
	})

	return initErr
}

// Close закрывает пул. Вызывать только при остановке процесса,
// а не в конце обработки запроса.
func Close() {
	if Pool != nil {
		Pool.Close()
	}
}
