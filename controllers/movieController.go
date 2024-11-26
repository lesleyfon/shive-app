package controllers

import (
	"context"
	"fmt"
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

func GetMovie() gin.HandlerFunc {
	return func(c *gin.Context) {

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var movie models.Movie

		movieId := c.Param("movie_id")

		movieFilter := bson.M{
			"movie_id": movieId,
		}

		fmt.Println(movieId)
		err := movieCollection.FindOne(ctx, movieFilter).Decode(&movie)

		if err != nil {
			c.JSON(
				http.StatusInternalServerError,
				gin.H{
					"status":  http.StatusInternalServerError,
					"message": "Error fetching movie from the db",
					"error":   err.Error(),
				},
			)
			return
		}
		c.JSON(
			http.StatusOK,
			gin.H{
				"status":  http.StatusOK,
				"data":    movie,
				"message": "Ok",
			},
		)
	}
}

func GetAllMovies() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)

		defer cancel()
		recordPerPageQueryKey := c.Query("recordPerPage")
		recordPerPage, err := strconv.Atoi(recordPerPageQueryKey)

		if err != nil || recordPerPage < 1 {
			recordPerPage = 10
		}

		pageQueryKey := c.Query("page")
		page, pageErr := strconv.Atoi(pageQueryKey)

		if pageErr != nil || page < 1 {
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

		groupStage := bson.D{
			{
				Key: "$group",
				Value: bson.D{
					{
						Key: "_id",
						Value: bson.D{
							{
								Key:   "_id",
								Value: "null",
							},
						},
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

		projectStage := bson.D{
			{
				Key: "$project",
				Value: bson.D{
					{
						Key:   "_id",
						Value: 0,
					}, {
						Key:   "total_count",
						Value: 1,
					}, {
						Key: "movie_items",
						Value: bson.D{{
							Key: "$slice",
							Value: []interface{}{
								"$data", startIndex, recordPerPage,
							}},
						},
					},
				},
			},
		}
		result, err := movieCollection.Aggregate(ctx, mongo.Pipeline{
			matchStage,
			groupStage,
			projectStage,
		})

		if err != nil {
			c.JSON(
				http.StatusInternalServerError,
				gin.H{
					"status":  http.StatusInternalServerError,
					"message": "error occurred while fetching all movies ",
					"error":   err.Error(),
				},
			)
			return
		}

		var allMovies []bson.M

		err = result.All(ctx, &allMovies)

		if err != nil {
			log.Fatal(err)
		}

		c.JSON(http.StatusOK,
			gin.H{
				"status":  http.StatusOK,
				"message": "Ok",
				"data":    allMovies[0],
			})
	}
}
