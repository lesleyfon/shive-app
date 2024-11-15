package routes

import (
	controller "shive/controllers"
	"shive/middleware"

	"github.com/gin-gonic/gin"
)

func GenreRouter(router *gin.Engine) {
	// Authenticates user
	router.Use(middleware.Authenticate())

	// Post route to create  a genre
	router.POST(
		"/genres/creategenre",
		controller.CreateGenre(),
	)

	router.GET(
		"/genres/:genre_id",
		controller.GetGenre(),
	)

	router.GET(
		"/genres",
		controller.GetAllGenres(),
	)
	router.PUT(
		"genres/:genre_id",
		controller.UpdateGenre(),
	)
}
