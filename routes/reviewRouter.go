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
	router.GET("/review/user_reviews/:reviewer_id", controllers.AllUserReviews())

	// DELETE Calls
	router.DELETE("/review/delete/:review_id", controllers.DeleteReviewByReviewId())

	// PUT Calls
	router.PUT("reviews/edit-review/:review_id", controllers.EditReviews())
}
