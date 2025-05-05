//local host settings
/*
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

	router.GET("/ping", handler.Ping)

	router.GET("/player/:id", handler.GetPlayer)

	router.GET("/matches", handler.GetMatches)

	router.POST("/matches", handler.PostMatches)

	router.Run()
}

func ErrRouter(c *gin.Context) {
	c.JSON(http.StatusBadRequest, gin.H{
		"errors": "this page could not be found",
	})
}
*/
package handler

import (
	"net/http"

	"fastcup/_pkg/routes"

	"github.com/gin-gonic/gin"
)

var (
	app *gin.Engine
)

// @title Golang Vercel Deployment
// @description API Documentation for Golang deployment in vercel serverless environment
// @version 1.0

// @schemes https http
// @host golang-vercel.vercel.app
func init() {
	app = gin.New()
	routes.Register(app)
}

// Entrypoint
func Handler(w http.ResponseWriter, r *http.Request) {
	app.ServeHTTP(w, r)
}
