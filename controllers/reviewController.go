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

var reviewCollection = database.OpenCollection(database.Client, "review")

func AddReview() gin.HandlerFunc {
	return func(c *gin.Context) {
		err := helpers.VerifyUserType(c, "USER")

		if err != nil {
			c.JSON(
				http.StatusUnauthorized,
				gin.H{
					"status":  http.StatusUnauthorized,
					"error":   err.Error(),
					"message": "Only users can add reviews",
				},
			)
			return
		}

		var review models.Review
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)

		defer cancel()

		//validate the request body
		err = c.BindJSON(&review)

		if err != nil {
			c.JSON(
				http.StatusBadRequest, gin.H{
					"status":  http.StatusBadRequest,
					"message": "error",
					"data":    err.Error(),
				},
			)
			return
		}

		//use the validator library to validate required fields
		validatorErr := validate.Struct(&review)
		if validatorErr != nil {
			c.JSON(
				http.StatusBadRequest, gin.H{
					"status":  http.StatusBadRequest,
					"message": "error",
					"data":    validatorErr.Error(),
				},
			)
			return
		}

		currentTime := time.Now()
		review.Id = primitive.NewObjectID()
		review.Review_id = review.Id.Hex()

		//REVIEW ID from REQUEST CONTEXT
		reviewId := c.GetString("uid")

		newReview := models.Review{
			Id:          review.Id,
			Created_at:  currentTime,
			Movie_id:    review.Movie_id,
			Review:      review.Review,
			Review_id:   review.Review_id,
			Reviewer_id: reviewId,
			Updated_at:  currentTime,
		}

		result, err := reviewCollection.InsertOne(ctx, newReview)

		if err != nil {
			c.JSON(
				http.StatusBadRequest,
				gin.H{
					"status":  http.StatusBadRequest,
					"message": "Error occurred while inserting a review into the database",
					"data":    err.Error(),
				},
			)
			return
		}

		c.JSON(
			http.StatusCreated,
			gin.H{
				"status":  http.StatusCreated,
				"message": "success",
				"data":    result,
			},
		)
	}
}

func GetAllMovieReviews() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*100)
		defer cancel()
		var reviews []models.Review
		movieId := c.Param("movie_id")

		if movieId == "" {
			c.JSON(
				http.StatusBadRequest,
				gin.H{
					"status":  http.StatusBadRequest,
					"message": "Ensure you add a movie Id to the endpoint",
				},
			)
			return
		}

		filter := bson.M{
			"movie_id": primitive.Regex{
				Pattern: movieId,
				Options: "i",
			},
		}

		searchedReviews, err := reviewCollection.Find(ctx, filter)

		if err != nil {
			c.JSON(
				http.StatusInternalServerError,
				gin.H{
					"status":  http.StatusInternalServerError,
					"error":   err.Error(),
					"message": "Error occurred while retrieving review data from the database",
				})
			return
		}

		err = searchedReviews.All(ctx, &reviews)

		if err != nil {
			c.JSON(
				http.StatusInternalServerError,
				gin.H{
					"status":  http.StatusInternalServerError,
					"error":   err.Error(),
					"message": "Error occurred while decoding filtered movie review data from the database",
				})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status":  http.StatusOK,
			"message": "filtered reviews!",
			"data":    reviews,
		})
	}
}
