package main

import (
	"log"
	"os"
	"shive/database"
	"shive/routes"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	port := os.Getenv("PORT")

	if port == "" {
		port = "8000"
	}
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}
	router := gin.Default()

	// run database
	database.StartDB()

	// LOG Events
	router.Use(gin.Logger())
	// Register app routes
	routes.AuthRoutes(router)
	routes.UserRoutes(router)
	routes.GenreRouter(router)
	routes.MovieRoutes(router)
	routes.ReviewRoutes(router)

	router.GET("/api", func(ctx *gin.Context) {
		ctx.JSON(
			200,
			gin.H{
				"success": "Welcome to Shive API!",
			},
		)
	})

	router.Run(":" + port)
}
