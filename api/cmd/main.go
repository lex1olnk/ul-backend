package main

import (
	"fastcup/_pkg/handler"

	"net/http"

	"github.com/gin-gonic/gin"
)

func HomepageHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Welcome to the Tech Company listing API with Golang"})
}

func main() {
	router := gin.Default()

	route := router.Group("/api")
	{
		route.GET("/ping", handler.Ping)

		route.GET("/player/:id", handler.GetPlayer)

		route.GET("/player/:id/matches", handler.GetPlayerMatchesByUlId)

		route.GET("/players", handler.GetPlayers)

		route.GET("/matches", handler.GetMatches)

		route.POST("/matches", handler.PostMatches)

		route.GET("/ultournaments", handler.GetUlTournaments)

		route.POST("/ultournaments", handler.PostUlTournaments)

		route.POST("/ulpicks", handler.PicksUlTournaments)

		route.POST("/ulmatches", handler.PostUlMatches)

		route.POST("/ulrating", handler.UpdateUlRating)
	}
	router.Run(":80")
}

func ErrRouter(c *gin.Context) {
	c.JSON(http.StatusBadRequest, gin.H{
		"errors": "this page could not be found",
	})
}
