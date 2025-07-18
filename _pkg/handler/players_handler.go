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

func GetPlayerMatchesByUlId(c *gin.Context) {
	// Получаем и валидируем параметры
	id := c.Param("id")
	ul_id := c.Query("ul_id")
	page := c.DefaultQuery("page", "0") // Значение по умолчанию для page

	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "player ID is required"})
		return
	}

	// Преобразование типов
	playerID, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid player ID format"})
		return
	}

	matchesPage, err := strconv.Atoi(page)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid page format"})
		return
	}

	// Подключение к базе данных
	if err := db.Init(); err != nil {
		log.Printf("Database connection error: %v", err)
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "service unavailable"})
		return
	}
	defer db.Close()

	// Контекст с таймаутом
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Выполнение запроса без транзакции (для чтения)
	matches, err := repository.InitialMatchData(ctx, db.Pool, playerID, ul_id, matchesPage)
	if err != nil {
		log.Printf("Database query error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch matches"})
		return
	}

	// Форматирование ответа
	response := gin.H{
		"status": "success",
		"data":   matches,
		"meta": gin.H{
			"player_id": playerID,
			"ul_id":     ul_id,
			"page":      matchesPage,
			"total":     len(matches),
			"timestamp": time.Now().Unix(),
		},
	}

	// Заголовки для предотвращения кеширования
	c.Header("Cache-Control", "no-store, max-age=0")
	c.JSON(http.StatusOK, response)
}

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

	playerTournaments, err := repository.GetPlayerUlTournaments(ctx, tx, playerID)
	if err != nil {
		log.Printf("Tournaments error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to ul tournaments stats"})
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
		"player_stats": avgStats,
		"maps_stats":   avgMapsStats,
		"tournaments":  playerTournaments,
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

	query := `SELECT player_id, nickname FROM players
ORDER BY nickname`

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
