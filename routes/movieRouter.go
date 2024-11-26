package routes

import (
	"shive/controllers"
	"shive/middleware"

	"github.com/gin-gonic/gin"
)

func MovieRoutes(router *gin.Engine) {
	// Auth middleware
	router.Use(middleware.Authenticate())

	router.POST("/movies/create-movie", controllers.CreateMovie())
}
