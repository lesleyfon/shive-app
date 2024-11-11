package routes

import (
	controllers "shive/controllers"

	"github.com/gin-gonic/gin"
)

func AuthRoutes(router *gin.Engine) {

	// Login
	router.POST("/users/login", controllers.Login())

	// Signup Route
	router.POST("/users/signup", controllers.Signup())
}
