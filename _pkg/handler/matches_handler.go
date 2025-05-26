package handler

import (
	"context"
	"fastcup/_pkg/db"
	"fastcup/_pkg/googleDocs"
	"fastcup/_pkg/repository"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

func GetMatches(c *gin.Context) {
	if err := db.Init(); err != nil {
		c.JSON(http.StatusExpectationFailed, gin.H{"Message": "failed connect to db"})
		return
	}
	defer db.Close()
	ctx := context.Background()
	players, err := repository.GetAggregatedPlayerStats(ctx, db.Pool)

	if err != nil {
		c.JSON(http.StatusExpectationFailed, gin.H{"Message": err})
		return
	}

	data := struct {
		Players []gin.H
	}{
		Players: players,
	}

	c.JSON(http.StatusOK, data)
}

func PostMatches(c *gin.Context) {
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

	resp, err := googleDocs.Init(c, ctx, "src!A1:A100")
	if err != nil {
		c.JSON(http.StatusExpectationFailed, gin.H{"Message": "failed fetch data"})
		return
	}

	re := regexp.MustCompile(`matches/(\d+)`)
	// 7. Проверяем и выводим данные
	if len(resp.Values) == 0 {
		c.JSON(http.StatusExpectationFailed, gin.H{"Message": "failed fetch excel data"})
	}

	fmt.Println("Полученные данные:")
	for _, row := range resp.Values {
		url := re.FindStringSubmatch(row[0].(string))
		matchID, err := strconv.Atoi(url[1])
		fmt.Println(matchID)
		if err != nil {
			// ... handle error
			c.JSON(http.StatusExpectationFailed, gin.H{"Message": "matchIdincorrect"})
			panic(err)
		}

		err = repository.CreateMatch(ctx, tx, matchID, nil)
		if err != nil {
			c.JSON(http.StatusExpectationFailed, gin.H{"Message": err})
			panic(err)
		}

	}
	// Фиксация транзакции
	if err := tx.Commit(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "transaction commit failed"})
		return
	}
	// Отправляем HTML-таблицу в ответе

	c.JSON(http.StatusOK, "OK")
}

type AttachMatchesRequest struct {
	TournamentID *string  `json:"tournament_id" binding:"omitempty,uuid"`
	MatchURLs    []string `json:"match_urls" binding:"required,min=1"`
}

func PostUlMatches(c *gin.Context) {
	var req AttachMatchesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid request data",
			"details": err.Error(),
		})
		return
	}

	// Извлекаем ID матчей из URL
	var matchIDs []int
	for _, url := range req.MatchURLs {
		parts := strings.Split(url, "/")
		if len(parts) < 1 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid match URL"})
			continue
		}

		matchID, err := strconv.Atoi(parts[len(parts)-1])
		if err != nil || matchID <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid match ID"})
			return
		}

		matchIDs = append(matchIDs, matchID)
	}

	tournamentId := req.TournamentID
	// Начало транзакции
	if err := db.Init(); err != nil {
		c.JSON(http.StatusExpectationFailed, gin.H{"Message": "failed connect to db"})
		return
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
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

	// Проверка существования турнира
	// Создание и привязка матчей
	for _, matchID := range matchIDs {
		// Проверка существования матча

		err = repository.CreateMatch(ctx, tx, matchID, tournamentId)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":    "failed to create match",
				"match_id": matchID,
				"message":  err.Error(),
			})
			return
		}

		if tournamentId != nil {
			// Привязка к турниру
			_, err := tx.Exec(ctx,
				`INSERT INTO tournament_matches(tournament_id, match_id)
				VALUES($1, $2)
				ON CONFLICT DO NOTHING`,
				req.TournamentID,
				matchID,
			)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":    "failed to attach match",
					"match_id": matchID,
				})
				return
			}
		}

	}

	// Фиксация транзакции
	if err := tx.Commit(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "transaction commit failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":        "success",
		"added_matches": len(matchIDs),
	})
}
