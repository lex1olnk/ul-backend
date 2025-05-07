package handler

import (
	"context"
	"fastcup/_pkg/db"
	"fastcup/_pkg/repository"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

func GetUlTournaments(c *gin.Context) {
	if err := db.Init(); err != nil {
		c.JSON(http.StatusExpectationFailed, gin.H{"Message": "failed connect to db"})
		return
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	tx, err := db.Pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "failed to begin transaction",
			"message": err.Error(), // Всегда используйте err.Error() для избежания сериализации
		})
		return
	}

	// Гарантируем откат/коммит транзакции
	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		}
	}()

	tournaments, err := repository.GetUlTournaments(ctx, tx)
	if err != nil {
		c.JSON(http.StatusExpectationFailed, gin.H{"Message": err.Error()})
		err = fmt.Errorf("get tournaments failed") // Помечаем для отката в defer
		return
	}

	// Коммитим транзакцию перед отправкой ответа
	if err = tx.Commit(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Message": "failed to commit transaction"})
		return
	}

	// Формируем ответ с данными
	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   tournaments, // Отправляем данные клиенту
	})
}

func PostUlTournaments(c *gin.Context) {
	if err := db.Init(); err != nil {
		c.JSON(http.StatusExpectationFailed, gin.H{"Message": "failed connect to db"})
		return
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	tx, err := db.Pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "failed to begin transaction",
			"message": err.Error(), // Всегда используйте err.Error() для избежания сериализации
		})
		return
	}

	// Гарантируем откат/коммит транзакции
	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		}
	}()

	name := c.PostForm("name")
	tournamentId, err := repository.PostUlTournaments(ctx, tx, name)

	if err != nil {
		c.JSON(http.StatusExpectationFailed, gin.H{"Message": err.Error()})
		err = fmt.Errorf("post tournaments failed") // Помечаем для отката в defer
		return
	}

	// Коммитим транзакцию перед отправкой ответа
	if err = tx.Commit(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Message": "failed to commit transaction"})
		return
	}

	// Формируем ответ с данными
	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": gin.H{
			"tournamentId": tournamentId,
		},
	})
}
