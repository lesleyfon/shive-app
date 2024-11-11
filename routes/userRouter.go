package routes

import (
	controllers "shive/controllers"
	"shive/middleware"

	"github.com/gin-gonic/gin"
)

func UserRoutes(router *gin.Engine) {

	// Auth middleware
	router.Use(middleware.Authenticate())
	// Get User
	router.GET("/users/:user_id", controllers.GetUser())
}
