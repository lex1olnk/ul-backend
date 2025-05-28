package handler

import (
	"context"
	"fastcup/_pkg/db"
	"fastcup/_pkg/googleDocs"
	"fastcup/_pkg/repository"
	"fmt"
	"net/http"
	"os"
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

func PicksUlTournaments(c *gin.Context) {
	name := c.PostForm("name")
	id := c.PostForm("id")
	if name == "" || id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "failed to get name",
		})
		return
	}
	fmt.Println(name)

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

	// Извлекаем N из строки вида "UMC#N"
	prefix := "UMC#"
	if !strings.HasPrefix(name, prefix) {
		// Обработка ошибки: неверный формат
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "postform have incorrect prefix",
			"message": err.Error(), // Всегда используйте err.Error() для избежания сериализации
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

	googleDocs.Init(c, ctx)
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

	spreadRange2 := "ulplayers!A2:D145"
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

	for _, row := range playersResp.Values {
		if len(row) < 3 {
			continue
		}

		nickname := strings.TrimSpace(row[0].(string))
		id, _ := strconv.Atoi(strings.TrimSpace(row[3].(string)))
		players[nickname] = id
	}

	// 7. Проверяем и выводим данные
	if len(resp.Values) == 0 {
		c.JSON(http.StatusExpectationFailed, gin.H{"Message": "failed fetch excel data"})
	}

	fmt.Println("Полученные данные:")
	for i, row := range resp.Values {
		for _, player := range row {
			//fmt.Println(i+1, player.(string), players[player.(string)])
			err = repository.PostUlPlayerPick(ctx, tx, players[player.(string)], id, i+1)
			if err != nil {
				c.JSON(http.StatusExpectationFailed, gin.H{"Message": err.Error()})
				return
			}
		}
	}

	for _, row := range winnersResp.Values {
		nickname := strings.TrimSpace(row[0].(string))
		err = repository.PostUlWinner(ctx, tx, players[nickname], id)
		if err != nil {
			c.JSON(http.StatusExpectationFailed, gin.H{"Message": err.Error()})
			return
		}
	}

	if err != nil {
		c.JSON(http.StatusExpectationFailed, gin.H{"Message": err.Error()})
		return
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
		"status": "success",
	})
}
