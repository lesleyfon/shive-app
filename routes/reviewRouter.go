package routes

import (
	"shive/controllers"
	"shive/middleware"

	"github.com/gin-gonic/gin"
)

func ReviewRoutes(router *gin.Engine) {
	// Auth Middleware
	router.Use(middleware.Authenticate())

	// POST Calls
	router.POST("/review/add-review", controllers.AddReview())

	// GET Calls
	router.GET("/review/filter/:movie_id", controllers.GetAllMovieReviews())
}
