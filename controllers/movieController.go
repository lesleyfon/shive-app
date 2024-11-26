package controllers

import (
	"context"
	"net/http"
	"shive/database"
	"shive/helpers"
	"shive/models"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var movieCollection = database.OpenCollection(database.Client, "movie")

func CreateMovie() gin.HandlerFunc {

	return func(c *gin.Context) {
		err := helpers.VerifyUserType(c, "ADMIN")
		if err != nil {
			c.JSON(
				http.StatusInternalServerError, gin.H{
					"status":  http.StatusInternalServerError,
					"message": "Error verifying user that is trying to create a movie",
					"error":   err.Error(),
				})
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		var movie models.Movie

		defer cancel()

		err = c.BindJSON(&movie)

		if err != nil {
			c.JSON(
				http.StatusInternalServerError, gin.H{
					"status":  http.StatusInternalServerError,
					"message": "error",
					"error":   err.Error(),
				})
			return
		}

		movieRegexMatch := bson.M{
			"name": bson.M{
				"$regex": primitive.Regex{
					Pattern: *movie.Name,
					Options: "i",
				},
			},
		}

		count, err := movieCollection.CountDocuments(ctx, movieRegexMatch)

		if err != nil {
			c.JSON(
				http.StatusInternalServerError,
				gin.H{
					"status":  http.StatusInternalServerError,
					"message": "Error occurred while checking if movie name exist in the db",
					"error":   err.Error(),
				})
			return
		}

		// Movie already exist
		if count > 0 {
			c.JSON(
				http.StatusBadRequest,
				gin.H{
					"status":  http.StatusBadRequest,
					"message": "error",
					"error":   "Movie already exist",
					"count":   count,
				})
			return
		}

		validationError := validate.Struct(&movie)
		if validationError != nil {

			c.JSON(
				http.StatusBadRequest,
				gin.H{
					"status":  http.StatusBadRequest,
					"message": "error",
					"error":   validationError.Error(),
				},
			)
			return
		}

		movie.Id = primitive.NewObjectID()
		movie.Movie_id = movie.Id.Hex()
		currentTime := time.Now()

		newMovie := models.Movie{
			Id:         movie.Id,
			Name:       movie.Name,
			Topic:      movie.Topic,
			Movie_id:   movie.Movie_id,
			Movie_URL:  movie.Movie_URL,
			Genre_id:   movie.Genre_id,
			Created_at: currentTime,
			Updated_at: currentTime,
		}

		result, err := movieCollection.InsertOne(ctx, newMovie)

		if err != nil {
			c.JSON(
				http.StatusInternalServerError,
				gin.H{
					"status":  http.StatusInternalServerError,
					"message": "Error occurred while creating a movie",
					"error":   err.Error(),
				},
			)
			return
		}
		// EVERY goes well, return with created
		c.JSON(
			http.StatusCreated, gin.H{
				"status":  http.StatusCreated,
				"message": "Movie Created",
				"data":    result,
			})
	}
}
