package handler

import (
	"context"
	"errors"
	"fastcup/_pkg/db"
	"fastcup/_pkg/repository"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

// GetPlayer возвращает детальную информацию об игроке
func GetPlayer(c *gin.Context) {
	// Валидация параметра
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "player ID is required"})
		return
	}

	playerID, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid player ID format"})
		return
	}
	if err := db.Init(); err != nil {
		c.JSON(http.StatusExpectationFailed, gin.H{"Message": "failed connect to db"})
		return
	}
	defer db.Close()
	// Инициализация контекста с таймаутом
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Начало транзакции
	tx, err := db.Pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		log.Printf("Transaction error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	defer func() {
		if err := tx.Rollback(ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			log.Printf("Rollback error: %v", err)
		}
	}()

	// Получение данных
	matches, err := repository.InitialMatchData(ctx, tx, playerID)
	if err != nil {
		log.Printf("Match data error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch matches"})
		return
	}

	avgStats, err := repository.GetAverageStats(ctx, tx, playerID)
	if err != nil {
		log.Printf("Avg stats error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch stats"})
		return
	}

	avgMapsStats, err := repository.GetAverageMapsStats(ctx, tx, playerID)
	if err != nil {
		log.Printf("Maps stats error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch map stats"})
		return
	}

	// Коммит транзакции
	if err := tx.Commit(ctx); err != nil {
		log.Printf("Commit error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	// Формирование ответа
	response := gin.H{
		"player_stats":   avgStats,
		"maps_stats":     avgMapsStats,
		"recent_matches": matches,
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   response,
	})
}

// GetPlayers возвращает список игроков
func GetPlayers(c *gin.Context) {
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

	query := `SELECT player_id, nickname FROM players`

	rows, err := tx.Query(ctx, query)
	if err != nil {
		log.Printf("Query error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	defer rows.Close()

	var players []Player
	for rows.Next() {
		var p Player
		if err := rows.Scan(&p.ID, &p.Nickname); err != nil {
			log.Printf("Scan error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}
		players = append(players, p)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Rows error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	if err = tx.Commit(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "transaction commit failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   gin.H{"players": players},
	})
}
