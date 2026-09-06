package handler

import (
	"context"
	"fastcup/_pkg/db"
	"fastcup/_pkg/googleDocs"
	"fastcup/_pkg/repository"
	"log"
	"net/http"
	"os"
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
	ctx := context.Background()
	players, err := repository.GetAggregatedPlayerStats(ctx, db.Pool, false, "")

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

	// Гарантируем откат: после успешного Commit Rollback безвреден (ErrTxClosed)
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	if err := googleDocs.Init(c, ctx); err != nil {
		// Init уже записал ответ об ошибке
		return
	}
	spreadsheetId := os.Getenv("GOOGLE_SHEET")

	resp, err := googleDocs.Srv.Spreadsheets.Values.Get(spreadsheetId, "src!A1:A100").Do()

	if err != nil {
		c.JSON(http.StatusExpectationFailed, gin.H{"Message": "failed fetch data"})
		return
	}

	re := regexp.MustCompile(`matches/(\d+)`)
	// 7. Проверяем и выводим данные
	if len(resp.Values) == 0 {
		c.JSON(http.StatusExpectationFailed, gin.H{"Message": "failed fetch excel data"})
		return
	}

	for _, row := range resp.Values {
		if len(row) == 0 {
			continue
		}

		match := re.FindStringSubmatch(cellString(row[0]))
		if match == nil {
			// Строка без ссылки вида matches/<id> — пропускаем
			continue
		}

		matchID, convErr := strconv.Atoi(match[1])
		if convErr != nil {
			err = convErr
			c.JSON(http.StatusExpectationFailed, gin.H{"Message": "match id is incorrect", "value": match[1]})
			return
		}

		if err = repository.CreateMatch(ctx, tx, matchID, nil); err != nil {
			c.JSON(http.StatusExpectationFailed, gin.H{"Message": err.Error(), "match_id": matchID})
			return
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

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	tx, err := db.Pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "failed to begin transaction",
			"message": err.Error(), // Всегда используйте err.Error() для избежания сериализации
		})
		return
	}

	// Гарантируем откат: после успешного Commit Rollback безвреден (ErrTxClosed)
	defer func() {
		_ = tx.Rollback(ctx)
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

func ExportMatchesByUlId(c *gin.Context) {
	ul_id := c.PostForm("id")
	ul_name := c.PostForm("name")
	if ul_name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "failed to get name",
		})
		return
	}

	if err := db.Init(); err != nil {
		c.JSON(http.StatusExpectationFailed, gin.H{"Message": "failed connect to db"})
		return
	}
	ctx := context.Background()

	if err := googleDocs.Init(c, ctx); err != nil {
		// Init уже записал ответ об ошибке
		return
	}
	srv := googleDocs.Srv
	spreadSheetId := os.Getenv("GOOGLE_SHEET")

	players, err := repository.GetAggregatedPlayerStats(ctx, db.Pool, true, ul_id)

	if err != nil {
		c.JSON(http.StatusExpectationFailed, gin.H{"Message": err})
		return
	}
	/* 	c.JSON(http.StatusOK, gin.H{
		"players": players,
	}) */

	if err := repository.WriteToGoogleSheets(ul_id, ul_name, players, srv, spreadSheetId); err != nil {
		log.Printf("Google Sheets error: %v", err)
	}

	c.JSON(http.StatusOK, gin.H{})
}
