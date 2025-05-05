package handler

import (
	"context"
	"fastcup/_pkg/db"
	"fastcup/_pkg/repository"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

func GetPlayer(c *gin.Context) {
	if err := db.Init(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to connect to database",
		})
		return
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Начало транзакции
	tx, err := db.Pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "failed to begin transaction",
			"message": err,
		})
		return
	}
	defer tx.Rollback(ctx) // Безопасный откат при ошибках

	// Получаем ID игрока из параметров запроса
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Player ID is required",
		})
		return
	}

	// Конвертируем ID в число (если в БД используется int)
	playerID, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid player ID format",
		})
		return
	}

	// Получаем последние 15 матчей игрока
	matches, err := repository.InitialMatchData(ctx, tx, playerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to fetch player matches",
			"message": err,
		})
		return
	}

	// Если нет матчей
	if len(matches) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"message": "No matches found for this player",
			"data":    nil,
		})
		return
	}

	// Рассчитываем средние значения
	avgStats, err := repository.GetAverageStats(ctx, tx, playerID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to fetch player avg stats",
			"message": err,
		})
		return
	}
	avgMapsStats, err := repository.GetAverageMapsStats(ctx, tx, playerID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to fetch player avg maps stats",
			"message": err,
		})
		return
	}

	// Формируем ответ
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

func GetPlayers(c *gin.Context) {
	if err := db.Init(); err != nil {
		c.JSON(http.StatusExpectationFailed, gin.H{"Message": "failed connect to db"})
		return
	}
	defer db.Close()
	ctx := context.Background()
	query := `SELECT player_id, nickname
	FROM players`

	rows, err := db.Pool.Query(ctx, query)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": err,
		})
		return
	}
	defer rows.Close()

	var players []Player
	for rows.Next() {
		var p Player
		if err := rows.Scan(
			&p.ID,
			&p.Nickname,
		); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status": "error",
			})
		}
		players = append(players, p)
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": gin.H{
			"players": players,
		},
	})
}
