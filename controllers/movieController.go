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

		// Find the movie by id
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

		// Get the record per page query
		recordPerPageQueryKey := c.Query("recordPerPage")
		recordPerPage, err := strconv.Atoi(recordPerPageQueryKey)

		// If the record per page query is not a number or is less than 1, set it to 10
		if err != nil || recordPerPage < 1 {
			recordPerPage = 10
		}

		// Get the page query
		pageQueryKey := c.Query("page")
		page, pageErr := strconv.Atoi(pageQueryKey)

		// If the page query is not a number or is less than 1, set it to 1
		if pageErr != nil || page < 1 {
			page = 1
		}

		// Calculate the start index
		startIndex := (page - 1) * recordPerPage

		// Match stage - Match the movie by id
		matchStage := bson.D{
			{
				Key:   "$match",
				Value: bson.D{{}},
			},
		}

		// Group stage - Group the movies by id
		groupStage := bson.D{
			{
				Key: "$group",
				Value: bson.D{
					// _id stage - Set the _id to null
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
						// Total count stage - Count the total number of movies
						Key: "total_count",
						Value: bson.D{{
							Key:   "$sum",
							Value: 1,
						}},
					},
					{
						// Data stage - Push the movies to the data array
						Key: "data",
						Value: bson.D{{
							Key:   "$push",
							Value: "$$ROOT",
						}},
					},
				},
			}}

		// Project stage - Project the movies
		projectStage := bson.D{
			{
				// Project stage - Project the movies
				Key: "$project",
				Value: bson.D{
					{
						// _id stage - Set the _id to 0
						Key:   "_id",
						Value: 0,
					}, {
						// Total count stage - Count the total number of movies
						Key:   "total_count",
						Value: 1,
					}, {
						// Movie items stage - Slice the movies
						Key: "movie_items",
						Value: bson.D{{
							// Slice stage - Slice the movies from the data array when we have pagination
							Key: "$slice",
							Value: []interface{}{
								"$data", startIndex, recordPerPage,
							}},
						},
					},
				},
			},
		}

		// Aggregate the movies
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

// Update a movie
func UpdateMovie() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

		movieId := c.Param("movie_id")

		var movie models.Movie

		defer cancel()

		objId, _ := primitive.ObjectIDFromHex(movieId)

		if err := c.BindJSON(&movie); err != nil {

			c.JSON(http.StatusBadRequest, gin.H{
				"status":  http.StatusBadRequest,
				"message": "error",
				"error":   err.Error(),
			})

			return
		}

		validationError := validate.Struct(&movie)
		if validationError != nil {

			c.JSON(
				http.StatusBadRequest,
				gin.H{
					"status":  http.StatusBadRequest,
					"message": "error validating movie data",
					"data":    validationError.Error(),
				},
			)
			return
		}

		update := bson.M{
			"name":       movie.Name,
			"topic":      movie.Topic,
			"genre_id":   movie.Genre_id,
			"movie_url":  movie.Movie_URL,
			"updated_at": time.Now(),
		}

		filterByID := bson.M{"_id": bson.M{"$eq": objId}}

		count, _ := movieCollection.CountDocuments(ctx, filterByID)

		if count < 1 {

			c.JSON(
				http.StatusBadRequest,
				gin.H{
					"status":  http.StatusBadRequest,
					"message": "No movie exist with the provided Id",
				},
			)
			return
		}

		result, err := movieCollection.UpdateOne(ctx, filterByID, bson.M{"$set": update})

		if err != nil {
			c.JSON(
				http.StatusInternalServerError,
				gin.H{
					"status":  http.StatusInternalServerError,
					"message": "error",
					"Data":    err.Error(),
				})
			return
		}

		var updatedMovie models.Movie

		if result.MatchedCount == 1 {

			err := movieCollection.FindOne(ctx, filterByID).Decode(&updatedMovie)

			if err != nil {
				c.JSON(
					http.StatusInternalServerError,
					gin.H{
						"status":  http.StatusInternalServerError,
						"message": "error",
						"error":   err.Error(),
					})
				return
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"status":  http.StatusOK,
			"message": "movie updated successfully!",
			"data":    updatedMovie,
		})
	}
}

func SearchMovieByQuery() gin.HandlerFunc {
	return func(c *gin.Context) {

		var searchedMovies []models.Movie

		movieName := c.Param("movieName")

		if movieName == "" {
			log.Println("movie name is empty")
			c.Header("Content-Type", "application/json")

			c.JSON(
				http.StatusNotFound,
				gin.H{
					"error": "Invalid search parameter" + movieName,
				},
			)
			c.Abort()
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		filter := bson.M{
			"name": bson.M{
				"$regex": primitive.Regex{
					Pattern: movieName,
					Options: "i",
				},
			},
		}
		searchQuery, err := movieCollection.Find(ctx, filter)

		if err != nil {
			c.JSON(
				http.StatusInternalServerError,
				gin.H{
					"status":  http.StatusInternalServerError,
					"error":   err.Error(),
					"message": "error occurred while searching for movie in db",
				},
			)
			return
		}

		err = searchQuery.All(ctx, &searchedMovies)

		if err != nil {
			c.JSON(
				http.StatusInternalServerError,
				gin.H{
					"status":  http.StatusInternalServerError,
					"error":   err.Error(),
					"message": "error occurred while decoding for movie in db",
				},
			)
			return
		}

		defer searchQuery.Close(ctx)
		err = searchQuery.Err()
		if err != nil {
			log.Println(err)
			c.IndentedJSON(
				http.StatusBadRequest,
				gin.H{
					"message": "invalid request",
					"status":  http.StatusBadRequest,
					"error":   err.Error(),
				},
			)
			return
		}
		defer cancel()
		c.IndentedJSON(200, searchedMovies)
	}
}
