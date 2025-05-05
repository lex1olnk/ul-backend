package main

import (
	"fastcup/_pkg/handler"
	"os"

	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/cors"
)

func HomepageHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Welcome to the Tech Company listing API with Golang"})
}

func main() {
	router := gin.Default()

	config := cors.Default()
	if os.Getenv("VERCEL_ENV") == "production" {
		config = cors.New(cors.Options{
			AllowedOrigins: []string{"https://vercel-fastcup.vercel.app/"},
		})
	} else {
		config = cors.AllowAll()
	}

	config.Handler(router)
	route := router.Group("/api")
	{
		route.GET("/ping", handler.Ping)

		route.GET("/player/:id", handler.GetPlayer)

		route.GET("/players", handler.GetPlayers)

		route.GET("/matches", handler.GetMatches)

		route.POST("/matches", handler.PostMatches)

		route.POST("/ulrating", handler.UpdateUlRating)
	}
	router.Run(":80")
}

func ErrRouter(c *gin.Context) {
	c.JSON(http.StatusBadRequest, gin.H{
		"errors": "this page could not be found",
	})
}
