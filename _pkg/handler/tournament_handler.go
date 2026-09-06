package handler

import (
	"context"
	"fastcup/_pkg/db"
	"fastcup/_pkg/googleDocs"
	"fastcup/_pkg/repository"
	"fmt"
	"net/http"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

func GetUlTournaments(c *gin.Context) {
	if err := db.Init(); err != nil {
		c.JSON(http.StatusExpectationFailed, gin.H{"Message": "failed connect to db"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	tournaments, err := repository.GetUlTournaments(ctx, db.Pool)
	if err != nil {
		c.JSON(http.StatusExpectationFailed, gin.H{"Message": err.Error()})
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

	name := c.PostForm("name")
	tournamentId, err := repository.PostUlTournaments(ctx, tx, name)

	if err != nil {
		c.JSON(http.StatusExpectationFailed, gin.H{"Message": err.Error()})
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

func PicksUlTournaments(c *gin.Context) {
	name := c.PostForm("name")
	id := c.PostForm("id")
	if name == "" || id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "failed to get name",
		})
		return
	}
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

	// Извлекаем N из строки вида "UMC#N"
	prefix := "UMC#"
	if !strings.HasPrefix(name, prefix) {
		// Обработка ошибки: неверный формат
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "postform have incorrect prefix",
			"message": fmt.Sprintf("name must start with %q, got %q", prefix, name),
		})
		return
	}

	nStr := strings.TrimPrefix(name, prefix)
	n, err := strconv.Atoi(nStr)
	if err != nil {
		// Обработка ошибки: N не является числом
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "failted to TrimPrefix",
			"message": err.Error(), // Всегда используйте err.Error() для избежания сериализации
		})
		return
	}

	if err := googleDocs.Init(c, ctx); err != nil {
		// Init уже записал ответ об ошибке
		return
	}
	spreadsheetId := os.Getenv("GOOGLE_SHEET")

	// Вычисляем границы диапазона
	start := 2 + (n-1)*5
	end := 2 + n*5 - 1

	// Формируем строку диапазона
	spreadRange := fmt.Sprintf("ОБЩАЯ ТАБЛИЦА!A%d:L%d", start, end)
	resp, err := googleDocs.Srv.Spreadsheets.Values.Get(spreadsheetId, spreadRange).Do()

	if err != nil {
		c.JSON(http.StatusExpectationFailed, gin.H{"Message": err.Error()})
		return
	}

	spreadRange2 := "ulplayers!A2:D200"
	playersResp, err := googleDocs.Srv.Spreadsheets.Values.Get(spreadsheetId, spreadRange2).Do()

	if err != nil {
		c.JSON(http.StatusExpectationFailed, gin.H{"Message": err.Error()})
		return
	}

	spreadRange3 := fmt.Sprintf("ОБЩАЯ ТАБЛИЦА!N%d:N%d", start, end)
	winnersResp, err := googleDocs.Srv.Spreadsheets.Values.Get(spreadsheetId, spreadRange3).Do()

	if err != nil {
		c.JSON(http.StatusExpectationFailed, gin.H{"Message": err.Error()})
		return
	}

	players := make(map[string]int)
	winners := []string{}

	for _, row := range playersResp.Values {
		if len(row) < 4 {
			continue
		}

		nickname := strings.TrimSpace(cellString(row[0]))
		playerID, convErr := strconv.Atoi(strings.TrimSpace(cellString(row[3])))
		if nickname == "" || convErr != nil {
			continue
		}
		players[nickname] = playerID
	}

	for _, row := range winnersResp.Values {
		if len(row) == 0 {
			continue
		}
		nickname := strings.TrimSpace(cellString(row[0]))
		winners = append(winners, nickname)
	}

	// 7. Проверяем и выводим данные
	if len(resp.Values) == 0 {
		c.JSON(http.StatusExpectationFailed, gin.H{"Message": "failed fetch excel data"})
		return
	}

	var unknown []string
	for i, row := range resp.Values {
		for _, cell := range row {
			nickname := strings.TrimSpace(cellString(cell))
			if nickname == "" {
				continue
			}

			playerID, ok := players[nickname]
			if !ok {
				// Не пишем пик на player_id = 0 для неизвестного ника
				unknown = append(unknown, nickname)
				continue
			}

			winner := slices.Contains(winners, nickname)
			if err = repository.PostUlPlayerPick(ctx, tx, playerID, id, i+1, winner); err != nil {
				c.JSON(http.StatusExpectationFailed, gin.H{"Message": err.Error()})
				return
			}
		}
	}

	// Коммитим транзакцию перед отправкой ответа
	if err = tx.Commit(ctx); err != nil {
		c.JSON(http.StatusInternalServerError,
			gin.H{
				"Message": "failed to commit transaction",
				"error":   err.Error(),
			})
		return
	}

	// Формируем ответ с данными
	c.JSON(http.StatusOK, gin.H{
		"status":           "success",
		"data":             resp.Values,
		"winners":          winners,
		"unknown_players":  unknown,
	})
}
