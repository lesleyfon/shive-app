package controllers

import (
	"context"
	"log"
	"net/http"
	"shive/database"
	"shive/helpers"
	"shive/models"
	"strconv"
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
				"error":   err.Error(),
				"message": "Error Validating Admin user",
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
				http.StatusBadRequest,
				gin.H{
					"status": http.StatusBadRequest,
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

func GetAllGenres() gin.HandlerFunc {
	return func(c *gin.Context) {
		userTypeValidationErr := helpers.VerifyUserType(c, "ADMIN")
		if userTypeValidationErr != nil {
			c.JSON(
				http.StatusBadRequest,
				gin.H{
					"status": http.StatusBadRequest,
					"error":  userTypeValidationErr.Error(),
				},
			)
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)

		defer cancel()

		recordPerPage, err := strconv.Atoi(c.Query("recordPerPage"))

		if err != nil || recordPerPage < 1 {
			recordPerPage = 10
		}

		page, err1 := strconv.Atoi("pages")
		if err1 != nil || page < 1 {
			page = 1
		}

		startIndex := (page - 1) * recordPerPage
		startIndex, err = strconv.Atoi(c.Query("startIndex"))

		matchStage := bson.D{
			{
				Key:   "$match",
				Value: bson.D{{}},
			},
		}
		// groupStage documents by a specified key and performs aggregate operations like counting, summing, etc.
		groupStage := bson.D{
			{
				Key: "$group",
				Value: bson.D{
					{
						Key: "_id",
						Value: bson.D{{
							Key:   "_id",
							Value: "null",
						}},
					},
					{
						Key: "total_count",
						Value: bson.D{{
							Key:   "$sum",
							Value: 1,
						}},
					},
					{
						Key: "data",
						Value: bson.D{{
							Key:   "$push",
							Value: "$$ROOT",
						}},
					},
				},
			}}
		// This stage reshapes the documents by including or excluding fields and performing transformations.
		projectStage := bson.D{
			{
				Key: "$project",
				Value: bson.D{
					{
						Key:   "_id",
						Value: 0,
					},
					{
						Key:   "total_count",
						Value: 1,
					},
					{
						Key: "genre_items",
						Value: bson.D{{
							Key: "$slice",
							Value: []interface{}{
								"$data",
								startIndex,
								recordPerPage,
							},
						}},
					}},
			}}
		result, err := genreCollection.Aggregate(ctx, mongo.Pipeline{
			matchStage,
			groupStage,
			projectStage,
		})

		defer cancel()
		if err != nil {
			c.JSON(
				http.StatusInternalServerError,
				gin.H{
					"status":  http.StatusInternalServerError,
					"message": "error occured while fetching genres ",
					"error":   err.Error(),
				},
			)
			return
		}

		var allGenres []bson.M

		err = result.All(ctx, &allGenres)
		if err != nil {
			log.Fatal(err)
		}

		c.JSON(http.StatusOK, allGenres[0])
	}

}

func UpdateGenre() gin.HandlerFunc {
	return func(c *gin.Context) {
		userAuthErr := helpers.VerifyUserType(c, "ADMIN")

		if userAuthErr != nil {
			c.JSON(http.StatusUnauthorized,
				gin.H{
					"status":  http.StatusUnauthorized,
					"error":   userAuthErr,
					"message": "Unauthorized",
				})
			return
		}

		genreId := c.Param("genre_id")
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)

		var genre models.Genre
		defer cancel()
		//	Validate request body
		err := c.BindJSON(&genre)
		if err != nil {
			c.JSON(
				http.StatusBadRequest,
				gin.H{
					"status": http.StatusBadRequest,
					"error":  err,
				})
			return
		}
		//use the validator library to validate required fields
		validationErr := validate.Struct(&genre)
		if validationErr != nil {
			c.JSON(
				http.StatusBadRequest,
				gin.H{
					"status":  http.StatusBadRequest,
					"error":   validationErr,
					"message": "Error validating request body",
				})
			return
		}

		update := bson.M{"name": genre.Name}
		filterById := bson.M{"genre_id": genreId}

		result, err := genreCollection.UpdateOne(
			ctx,
			filterById,
			bson.M{"$set": update},
		)

		if err != nil {
			c.JSON(
				http.StatusInternalServerError,
				gin.H{
					"status":  http.StatusInternalServerError,
					"message": "Error updating genre",
					"error":   err.Error(),
				})
			return
		}

		var updatedGenre models.Genre

		if result.MatchedCount == 1 {
			err := genreCollection.FindOne(ctx, filterById).Decode(&updatedGenre)

			if err != nil {
				c.JSON(
					http.StatusInternalServerError,
					gin.H{
						"status":  http.StatusInternalServerError,
						"error":   err,
						"message": "Error Retrieving updated genre from collection",
					})
				return
			}
		}
		c.JSON(
			http.StatusOK,
			gin.H{
				"Status":  http.StatusOK,
				"Message": "success",
				"Data":    updatedGenre,
			},
		)
	}
}

func DeleteGenre() gin.HandlerFunc {
	return func(c *gin.Context) {
		userAuthErr := helpers.VerifyUserType(c, "ADMIN")

		if userAuthErr != nil {
			c.JSON(
				http.StatusUnauthorized,
				gin.H{
					"status":  http.StatusUnauthorized,
					"message": "Unauthorized",
					"error":   userAuthErr.Error(),
				},
			)
			return
		}

		genreId := c.Param("genre_id")

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)

		defer cancel()

		result, err := genreCollection.DeleteOne(
			ctx,
			bson.M{
				"genre_id": genreId,
			},
		)

		if err != nil {
			c.JSON(
				http.StatusBadRequest,
				gin.H{
					"status":  http.StatusBadRequest,
					"error":   err.Error(),
					"message": "Error Deleting Genre",
				})
			return
		}

		if result.DeletedCount < 1 {
			c.JSON(http.StatusNotFound,
				gin.H{
					"status":  http.StatusNotFound,
					"message": "error",
					"data": map[string]interface{}{
						"data":     "Genre with specified ID not found!",
						"genre_id": genreId,
					},
				},
			)
			return
		}

		c.JSON(
			http.StatusOK,
			gin.H{
				"status":  http.StatusOK,
				"message": "success",
				"data": map[string]interface{}{
					"data": "Genre successfully deleted!",
				},
			},
		)
	}
}

func SearchByName() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var genres []models.Genre
		searchGenreName := c.Query("genre-name")

		if searchGenreName == "" {
			c.JSON(
				http.StatusBadRequest,
				gin.H{
					"error": "Please provide a genre name to search",
				})
			return
		}

		partialMatchSearchRegexMatch := bson.M{
			"name": bson.M{
				"$regex": primitive.Regex{
					Pattern: searchGenreName,
					Options: "i", // Case-insensitive
				},
			},
		}

		// Execute the query to find genres
		cursor, err := genreCollection.Find(ctx, partialMatchSearchRegexMatch)

		if err != nil {
			c.JSON(
				http.StatusInternalServerError,
				gin.H{
					"error": "Error occurred while querying genres",
				})
			return
		}

		err = cursor.All(ctx, &genres)

		if err != nil {
			c.JSON(
				http.StatusInternalServerError,
				gin.H{
					"error": err.Error(),
				})
			return
		}

		if len(genres) == 0 {
			c.JSON(
				http.StatusNotFound,
				gin.H{
					"message": "No genres found matching the search criteria",
				})
			return
		}

		c.JSON(
			http.StatusOK,
			gin.H{
				"status":          http.StatusOK,
				"searchGenreName": searchGenreName,
				"data":            genres,
			})
	}
}
