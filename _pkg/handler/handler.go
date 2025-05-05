package handler

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"fastcup/_pkg/db"

	"github.com/gin-gonic/gin"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

func Ping(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"ping": "pong"})
}

// @Tags        Welcome
// @Summary     Hello User
// @Description Endpoint to Welcome user and say Hello "Name"
// @Param       name query string true "Name in the URL param"
// @Accept      json
// @Produce     json
// @Success     200 {object} object "success"
// @Failure     400 {object} object "Request Error or parameter missing"
// @Failure     404 {object} object "When user not found"
// @Failure     500 {object} object "Server Error"
// @Router      /hello/:name [GET]

// GraphQLRequest структура для GraphQL-запроса
//

// MatchHandler обрабатывает запросы к маршруту /match/{id}

type Player struct {
	ID       int    `json:"id"`
	Nickname string `json:"nickname"`
}

func UpdateUlRating(c *gin.Context) {
	if err := db.Init(); err != nil {
		c.JSON(http.StatusExpectationFailed, gin.H{"Message": "failed connect to db"})
		return
	}
	defer db.Close()

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

	ctx := context.Background()
	// 3. Создаем сервис Sheets с учетными данными из файла
	srv, err := sheets.NewService(ctx, option.WithCredentialsJSON([]byte(googleCreds)))
	if err != nil {
		c.JSON(http.StatusExpectationFailed, gin.H{"Message": "failed connect to google sheet"})
		return
	}

	// 4. ID документа (из URL Google Sheets)
	spreadsheetId := os.Getenv("GOOGLE_SHEET")

	// 5. Диапазон для чтения, например, "Sheet1!A1:C10"
	playersRange := "ulplayers!A2:D300"
	playersResp, err := srv.Spreadsheets.Values.Get(spreadsheetId, playersRange).Do()

	if err != nil || len(playersResp.Values) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to get players data", "msg": playersResp})
		return
	}

	// Создаем мапы для связи данных
	type PlayerData struct {
		ID       int
		UlRating float64
		Faceit   string
	}
	players := make(map[string]PlayerData)
	for _, row := range playersResp.Values {
		if len(row) < 3 {
			continue
		}

		nickname := strings.TrimSpace(row[0].(string))
		faceit := strings.TrimSpace(row[1].(string))
		id, _ := strconv.Atoi(strings.TrimSpace(row[3].(string)))

		players[nickname] = PlayerData{
			ID:       id,
			UlRating: 0,
			Faceit:   faceit,
		}
	}

	// 2. Получаем рейтинги из листа с UL рейтингами
	ratingsRange := "ulrating!B1:FF" // A: ник, B: UL рейтинг
	ratingsResp, err := srv.Spreadsheets.Values.Get(spreadsheetId, ratingsRange).Do()
	if err != nil || len(ratingsResp.Values) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to get ratings data"})
		return
	}

	for i := range ratingsResp.Values[0] {
		nickname := strings.TrimSpace(ratingsResp.Values[0][i].(string))
		ulratingstr := strings.Replace(strings.TrimSpace(ratingsResp.Values[1][i].(string)), ",", ".", 1)
		ulrating, _ := strconv.ParseFloat(ulratingstr, 64)

		// If the key exists
		if p, ok := players[nickname]; ok {
			p.UlRating = ulrating
			players[nickname] = p
		}
	}

	queryUpdate := `
		UPDATE players 
		SET 
				ul_rating = $1,
				faceit = CASE 
						WHEN faceit IS NULL OR faceit = '' THEN $2 
						ELSE faceit 
				END
		WHERE player_id = $3
	`

	for _, player := range players {
		var exists bool
		err = db.Pool.QueryRow(ctx,
			"SELECT EXISTS(SELECT 1 FROM players WHERE player_id = $1)",
			player.ID).Scan(&exists)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}

		if exists {
			_, err := db.Pool.Exec(ctx, queryUpdate, player.UlRating, player.Faceit, player.ID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err})
				return
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": players})
}
