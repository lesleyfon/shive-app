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
	"go.mongodb.org/mongo-driver/mongo"
)

var genreCollection *mongo.Collection = database.OpenCollection(database.Client, "genre")

func CreateGenre() gin.HandlerFunc {
	return func(c *gin.Context) {

		if err := helpers.VerifyUserType(c, "ADMIN"); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		var genre models.Genre
		defer cancel()

		err := c.BindJSON(&genre)

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  http.StatusBadRequest,
				"message": "error",
				"error":   err.Error(),
			})
			return
		}

		regExpMach := bson.M{
			"$regex": primitive.Regex{
				Pattern: *genre.Name,
				Options: "i",
			},
		}
		count, err := genreCollection.CountDocuments(ctx, bson.M{
			"Name": regExpMach,
		})

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  http.StatusBadRequest,
				"message": "error occured while checking for the genre name",
				"data":    err.Error(),
			})
			return
		}
		if count > 0 {
			c.JSON(
				http.StatusUnauthorized,
				gin.H{
					"status": http.StatusUnauthorized,
					"error":  "this genre name already exists",
					"count":  count,
				},
			)
		}

		if validationError := validate.Struct(&genre); validationError != nil {

			c.JSON(http.StatusBadRequest,
				gin.H{
					"status":  http.StatusBadRequest,
					"message": "error",
					"error":   validationError.Error(),
				},
			)
			return
		}
		genre.Id = primitive.NewObjectID()
		genre.Genre_id = genre.Id.Hex()

		newGenre := models.Genre{
			Id:         genre.Id,
			Name:       genre.Name,
			Genre_id:   genre.Genre_id,
			Created_at: time.Now(),
			Updated_at: time.Now(),
		}

		result, err := genreCollection.InsertOne(ctx, newGenre)

		if err != nil {

			c.JSON(
				http.StatusBadRequest,
				gin.H{
					"status":  http.StatusBadRequest,
					"message": "error",
					"error":   err.Error(),
				},
			)
			return
		}

		c.JSON(
			http.StatusCreated,
			gin.H{
				"status":  http.StatusCreated,
				"message": "genre created successfully",
				"data":    result,
			},
		)
	}
}

func GetGenre() gin.HandlerFunc {
	return func(c *gin.Context) {

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)

		defer cancel()

		var genre models.Genre

		// Get genre Id from request url
		genreId := c.Param("genre_id")

		filter := bson.M{
			"genre_id": genreId,
		}

		// Get item from collection
		err := genreCollection.FindOne(ctx, filter).Decode(&genre)

		if err != nil {
			c.JSON(
				http.StatusInternalServerError,
				gin.H{
					"status":  http.StatusInternalServerError,
					"message": "Error retrieving genre",
					"error":   err.Error(),
				},
			)
			return
		}
		c.JSON(
			http.StatusAccepted,
			gin.H{
				"status": http.StatusAccepted,
				"data":   genre,
			},
		)
	}
}
