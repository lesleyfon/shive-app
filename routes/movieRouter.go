package routes

import (
	"shive/controllers"
	"shive/middleware"

	"github.com/gin-gonic/gin"
)

func MovieRoutes(router *gin.Engine) {
	// Auth middleware
	router.Use(middleware.Authenticate())
	// POST calls
	router.POST("/movies/create-movie", controllers.CreateMovie())

	// GET Calls
	router.GET("/movies/:movie_id", controllers.GetMovie())
	router.GET("/movies", controllers.GetAllMovies())
	router.GET("/movies/search/:movieName", controllers.SearchMovieByQuery())
	router.GET("/movies/filter/:genreId", controllers.SearchMovieByGenreId())

	router.PUT("/movies/:movie_id", controllers.UpdateMovie())
}
