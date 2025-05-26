package googleDocs

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

func Init(c *gin.Context, ctx context.Context, spreadRange string) (*sheets.ValueRange, error) {
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
		return nil, err
	}

	// 4. ID документа (из URL Google Sheets)
	spreadsheetId := os.Getenv("GOOGLE_SHEET")

	// 5. Диапазон для чтения, например, "Sheet1!A1:C10"

	// 6. Получаем значения указываем диапазон
	resp, err := srv.Spreadsheets.Values.Get(spreadsheetId, spreadRange).Do()

	return resp, err
}
