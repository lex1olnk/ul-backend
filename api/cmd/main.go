package main

import (
	"fastcup/_pkg/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()

	// Тот же набор маршрутов, что и в serverless-точке входа
	routes.Register(router)

	router.Run(":5000")
}
