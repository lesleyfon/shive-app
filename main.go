package main

import (
	"fmt"
	"os"
	"shive/database"
	"shive/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	port := os.Getenv("PORT")

	if port == "" {
		port = "8000"
	}

	router := gin.Default()

	// run database
	database.StartDB()

	// LOG Events
	router.Use(gin.Logger())

	// Register app routes
	routes.AuthRoutes(router)
	routes.UserRoutes(router)

	router.GET("/api", func(ctx *gin.Context) {
		ctx.JSON(
			200,
			gin.H{
				"success": "Welcome to Shive API!",
			},
		)
	})

	router.Run(":" + port)
	fmt.Print("Hello World")
}
