package main

import (
	"github.com/fatorin/mmr-tracker/config"
	"github.com/fatorin/mmr-tracker/database"
	"github.com/fatorin/mmr-tracker/routes"
	"github.com/gin-gonic/gin"
)

func main() {
	config.LoadEnv()
	database.Connect()

	r := gin.Default()
	routes.RegisterRoutes(r)

	r.Run(":8080")
}
