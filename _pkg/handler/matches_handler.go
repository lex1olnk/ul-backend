package handler

import (
	"context"
	"fastcup/_pkg/db"
	"fastcup/_pkg/repository"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
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

	googleCreds := fmt.Sprintf(`{
		"type": "service_account",
		"project_id": "%s",
		"private_key_id": "%s",
		"private_key": "%s",
		"client_email": "%s",
		"client_id": "%s",
		"project_id": "%s",		
		"auth_uri": "https://accounts.google.com/o/oauth2/auth",
		"token_uri": "https://oauth2.googleapis.com/token",
		"auth_provider_x509_cert_url": "https://www.googleapis.com/oauth2/v1/certs",
		"client_x509_cert_url": "%s",
		"universe_domain": "googleapis.com"
	}`,
		os.Getenv("GOOGLE_PROJECT_ID"),
		os.Getenv("GOOGLE_PRIVATE_KEY_ID"),
		os.Getenv("GOOGLE_PRIVATE_KEY"),
		os.Getenv("GOOGLE_CLIENT_EMAIL"),
		os.Getenv("GOOGLE_CLIENT_ID"),
		os.Getenv("GOOGLE_PROJECT_ID"),
		os.Getenv("GOOGLE_CLIENT_X509_CERT_URL"),
	)

	// 3. Создаем сервис Sheets с учетными данными из файла
	srv, err := sheets.NewService(ctx, option.WithCredentialsJSON([]byte(googleCreds)))
	if err != nil {
		c.JSON(http.StatusExpectationFailed, gin.H{"Message": "failed connect to google sheet"})
		return
	}

	// 4. ID документа (из URL Google Sheets)
	spreadsheetId := os.Getenv("GOOGLE_SHEET")

	// 5. Диапазон для чтения, например, "Sheet1!A1:C10"
	readRange := "src!A1:A100"

	// 6. Получаем значения указываем диапазон
	resp, err := srv.Spreadsheets.Values.Get(spreadsheetId, readRange).Do()
	if err != nil {
		c.JSON(http.StatusExpectationFailed, gin.H{"Message": "failed fetch data"})
	}

	re := regexp.MustCompile(`matches/(\d+)`)
	// 7. Проверяем и выводим данные
	if len(resp.Values) == 0 {
		c.JSON(http.StatusExpectationFailed, gin.H{"Message": "failed fetch excel data"})
	}

	fmt.Println("Полученные данные:")
	fmt.Println(spreadsheetId)
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
			return
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
