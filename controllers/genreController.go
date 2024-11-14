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
				"Status":  http.StatusBadRequest,
				"Message": "error",
				"Data": map[string]interface{}{
					"data": err.Error(),
				}})
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
				"Status":  http.StatusBadRequest,
				"Message": "error occured while checking for the genre name",
				"Data": map[string]interface{}{
					"data": err.Error(),
				},
			})
			return
		}
		if count > 0 {
			c.JSON(
				http.StatusUnauthorized,
				gin.H{
					"Status": http.StatusUnauthorized,
					"Error":  "this genre name already exists",
					"Count":  count,
				},
			)
		}

		if validationError := validate.Struct(&genre); validationError != nil {

			c.JSON(http.StatusBadRequest,
				gin.H{
					"Status":  http.StatusBadRequest,
					"Message": "error",
					"Data": map[string]interface{}{
						"data": validationError.Error(),
					},
				},
			)
			return
		}

		newGenre := models.Genre{
			Id:         primitive.NewObjectID(),
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
					"Status":  http.StatusBadRequest,
					"Message": "error",
					"Data": map[string]interface{}{
						"data": err.Error(),
					},
				},
			)
			return
		}

		c.JSON(
			http.StatusCreated,
			gin.H{
				"Status":  http.StatusCreated,
				"Message": "genre created successfully",
				"Data": map[string]interface{}{
					"data": result,
				},
			},
		)
	}
}
