package routes

import (
	"fastcup/_pkg/handler"
	"fastcup/_pkg/middleware"

	"net/http"

	"github.com/gin-gonic/gin"
)

func Register(app *gin.Engine) {
	// gin.New() идёт без middleware, поэтому Recovery добавляем явно:
	// паника в хендлере иначе убивает весь инстанс
	app.Use(gin.Recovery())
	app.NoRoute(ErrRouter)

	route := app.Group("/api")
	{
		// Публичные (чтение)
		route.GET("/ping", handler.Ping)

		route.GET("/player/:id", handler.GetPlayer)

		route.GET("/players", handler.GetPlayers)

		route.GET("/player/:id/matches", handler.GetPlayerMatchesByUlId)

		route.GET("/matches", handler.GetMatches)

		route.GET("/ultournaments", handler.GetUlTournaments)
	}

	// Мутирующие эндпоинты — только с валидным API_TOKEN
	write := app.Group("/api", middleware.RequireAPIToken())
	{
		write.POST("/matches", handler.PostMatches)

		write.POST("/matches/export", handler.ExportMatchesByUlId)

		write.POST("/ultournaments", handler.PostUlTournaments)

		write.POST("/ulpicks", handler.PicksUlTournaments)

		write.POST("/ulmatches", handler.PostUlMatches)

		write.POST("/ulrating", handler.UpdateUlRating)
	}
}

func ErrRouter(c *gin.Context) {
	c.JSON(http.StatusNotFound, gin.H{
		"errors": "this page could not be found",
	})
}
