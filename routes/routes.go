package routes

import (
	"os"

	"github.com/fatorin/mmr-tracker/handlers"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine) {
	r.Static("/assets", "./assets")
	r.LoadHTMLGlob("templates/*")
	r.GET("/api/scores", handlers.GetScores)
	r.GET("/api/match_histories", handlers.GetMatchHistories)
	r.GET("/", func(c *gin.Context) {
		c.HTML(200, "index.html", nil)
	})
	r.GET("/match_history", func(c *gin.Context) {
		c.HTML(200, "match_history.html", nil)
	})
	r.Use(cors.New(getCORSConfig()))
}

func getCORSConfig() cors.Config {
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = []string{os.Getenv("CORS")}
	corsConfig.AllowMethods = []string{"GET", "OPTIONS"}
	corsConfig.AllowHeaders = []string{"Origin", "Content-Type", "Accept"}
	corsConfig.AllowCredentials = true
	return corsConfig
}
