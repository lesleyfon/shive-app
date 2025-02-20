package controllers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"shive/database"
	helper "shive/helpers"
	"shive/models"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

var userCollection *mongo.Collection = database.OpenCollection(database.Client, "user")
var validate = validator.New()

// The function `MaskPassword` generates a bcrypt hash from a given password.
func MaskPassword(password string) string {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		log.Panic(err)
	}
	return string(bytes)
}

func Signup() gin.HandlerFunc {

	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		var user models.User

		defer cancel()

		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"Status":  http.StatusBadRequest,
				"Message": "error",
				"Data":    map[string]interface{}{"data": err.Error()}})
			return
		}

		//Check to see if name exists
		regexMatch := bson.M{"$regex": primitive.Regex{Pattern: *user.Email, Options: "i"}}
		emailCount, emailErr := userCollection.CountDocuments(ctx, bson.M{"email": regexMatch})
		usernameMatch := bson.M{"$regex": primitive.Regex{Pattern: *user.Username, Options: "i"}}
		usernameCount, usernameErr := userCollection.CountDocuments(ctx, bson.M{"username": usernameMatch})

		defer cancel()
		if emailErr != nil {
			log.Panic(emailErr)
			c.JSON(http.StatusBadRequest, gin.H{"error": "error occurred while checking for this email"})
		}
		if emailCount > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Looks like this email already exists", "count": emailCount})
			return
		}
		if usernameErr != nil {
			log.Panic(usernameErr)
			c.JSON(http.StatusBadRequest, gin.H{"error": "error occurred while checking for this email / username"})
		}
		if usernameCount > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Looks like this username already exists", "count": usernameCount})
			return
		}

		//To hash the password before sending it to the db
		password := MaskPassword(*user.Password)
		user.Password = &password
		user.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.ID = primitive.NewObjectID()
		user.User_id = user.ID.Hex()

		//Sign details to token
		token, refreshToken, _ := helper.GenerateAllTokens(
			*user.Email,
			*user.Name,
			*user.Username,
			*user.User_type,
			*&user.User_id)
		user.Token = &token
		user.Refresh_token = &refreshToken

		//Check to see if data being passed meets the requirements
		if validationError := validate.Struct(&user); validationError != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"Status":  http.StatusBadRequest,
				"Message": "error",
				"Data":    map[string]interface{}{"data": validationError.Error()}})
			return
		}

		if validationError := validate.Struct(&user); validationError != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"Status":  http.StatusBadRequest,
				"Message": "error",
				"Data":    map[string]interface{}{"data": validationError.Error()}})
			return
		}

		//To add a new user to the database
		newUser := models.User{
			ID:            user.ID,
			User_id:       user.ID.Hex(),
			Name:          user.Name,
			Username:      user.Username,
			Email:         user.Email,
			Password:      user.Password,
			Created_at:    user.Created_at,
			Updated_at:    user.Updated_at,
			Token:         user.Token,
			User_type:     user.User_type,
			Refresh_token: user.Refresh_token,
		}

		_, err := userCollection.InsertOne(ctx, newUser)

		//Error messages
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"Status":  http.StatusBadRequest,
				"Message": "error",
				"Data":    map[string]interface{}{"data": err.Error()}})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"status":  http.StatusCreated,
			"message": "User created successfully!",
			"data": map[string]string{
				"user_id":       newUser.User_id,
				"token":         *newUser.Token,
				"refresh_token": *newUser.Refresh_token,
				"name":          *newUser.Name,
				"username":      *newUser.Username,
				"email":         *newUser.Email,
				"user_type":     *newUser.User_type,
			},
		})
	}

}

func ConfirmPassword(userPassword string, passwordEntered string) (bool, string) {
	err := bcrypt.CompareHashAndPassword([]byte(passwordEntered), []byte(userPassword))
	check := true
	msg := ""

	if err != nil {
		msg = fmt.Sprintf("Looks like you entered a wrong password")
		check = false
	}

	return check, msg

}

func Login() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

		var user models.User
		var retrievedUser models.User

		defer cancel()

		// This block of code is attempting to bind the JSON data from the request body to the `user` struct.
		// If there is an error during this process (e.g., the JSON data cannot be properly bound to the
		// struct), it will return a response with a status code of `http.StatusInternalServerError` and a
		// JSON object containing an error message indicating that the email or password is incorrect.
		if err := c.BindJSON(&user); err != nil {
			c.JSON(
				http.StatusBadRequest,
				gin.H{
					"message": "Error occurred while binding JSON",
					"error":   err.Error(),
				},
			)
		}

		// This line of code is querying the `userCollection` (which is a MongoDB collection) to find a
		// document that matches the specified filter criteria. In this case, it is looking for a document
		// where the value of the "email" field matches the email provided in the `user` struct.
		err := userCollection.FindOne(ctx, bson.M{
			"email": user.Email,
		}).Decode(
			&retrievedUser,
		)

		if err != nil {
			c.JSON(
				http.StatusBadRequest,
				gin.H{
					"message": "Error occurred while finding user",
					"error":   err.Error(),
				},
			)
			return
		}

		passwordIsValid, msg := ConfirmPassword(*user.Password, *retrievedUser.Password)

		defer cancel()
		if !passwordIsValid {
			c.JSON(
				http.StatusBadRequest,
				gin.H{
					"message": msg,
					"error":   err.Error(),
				},
			)
		}

		if retrievedUser.Email == nil {
			c.JSON(
				http.StatusBadRequest,
				gin.H{
					"message": "Oops account not found",
					"error":   err.Error(),
				},
			)
		}

		token, refreshedToken, _ := helper.GenerateAllTokens(
			*user.Email,
			*user.Name,
			*retrievedUser.Username,
			*user.User_type,
			retrievedUser.User_id,
		)

		updatedUser, err := helper.UpdateTokens(token, refreshedToken, user.User_id)
		// TODO: get the details from updateTokens method
		updatedUser.Email = retrievedUser.Email
		updatedUser.Name = retrievedUser.Name
		updatedUser.Username = retrievedUser.Username
		updatedUser.User_id = retrievedUser.User_id
		updatedUser.User_type = retrievedUser.User_type

		defer cancel()

		if err != nil {
			c.JSON(
				http.StatusInternalServerError,
				gin.H{
					"message": "Error occurred while updating tokens",
					"error":   err.Error(),
				},
			)
			return
		}

		c.JSON(
			http.StatusOK,
			updatedUser,
		)
	}
}

func GetUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		userId := c.Param("user_id")

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

		var user models.User

		err := userCollection.FindOne(ctx, bson.M{
			"user_id": userId,
		}).Decode(&user)

		defer cancel()

		if err != nil {
			c.JSON(
				http.StatusInternalServerError,
				gin.H{
					"error": err.Error(),
				},
			)
		}

		c.JSON(
			http.StatusOK,
			user,
		)
	}
}

func GetUsers() gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := helper.VerifyUserType(c, "ADMIN"); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		recordPerPage, err := strconv.Atoi(c.Query("recordPerPage"))

		if err != nil || recordPerPage < 1 {
			recordPerPage = 10
		}

		page, err1 := strconv.Atoi(c.Query("page"))

		if err1 != nil || page < 1 {
			page = 1
		}

		startIndex := (page - 1) * recordPerPage
		startIndex, err = strconv.Atoi(c.Query("startIndex"))

		matchStage := bson.D{{Key: "$match", Value: bson.D{{}}}}
		groupStage := bson.D{
			{
				Key: "$group", Value: bson.D{
					{
						Key: "_id", Value: bson.D{
							{
								Key: "_id", Value: "null",
							},
						},
					},
					{
						Key: "total_count",
						Value: bson.D{
							{
								Key: "$sum", Value: 1,
							},
						},
					},
					{
						Key: "data", Value: bson.D{
							{
								Key:   "$push",
								Value: "$$ROOT",
							},
						},
					},
				},
			},
		}
		projectStage := bson.D{
			{
				Key: "$project", Value: bson.D{
					{
						Key: "_id", Value: 0,
					},
					{
						Key: "total_count", Value: 1,
					},
					{
						Key: "user_items", Value: bson.D{
							{
								Key:   "$slice",
								Value: []interface{}{"$data", startIndex, recordPerPage},
							},
						},
					},
				},
			},
		}
		result, err := userCollection.Aggregate(ctx, mongo.Pipeline{
			matchStage, groupStage, projectStage})

		defer cancel()
		if err != nil {
			c.JSON(
				http.StatusInternalServerError,
				gin.H{
					"error": "error occurred while listing user items",
				},
			)
		}
		var allusers []bson.M
		if err = result.All(ctx, &allusers); err != nil {
			log.Fatal(err)
		}
		c.JSON(http.StatusOK, allusers[0])
	}
}

func CreateUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		var user models.User

		defer cancel()

		err := userCollection.FindOne(ctx, bson.M{
			"email": user.Email,
		}).Decode(
			&user,
		)

		if err != nil {
			c.JSON(
				http.StatusBadRequest,
				gin.H{
					"error": "An Error occurred",
				},
			)
			return
		}

		if user.Email != nil {
			c.JSON(
				http.StatusBadRequest,
				gin.H{
					"error": "User already Exist",
				},
			)
		}
		regexpMatch := bson.M{
			"$regex": primitive.Regex{
				Pattern: *user.Email,
				Options: "i",
			},
		}
		emailCount, emailErr := userCollection.CountDocuments(ctx, bson.M{"email": regexpMatch})

		if emailErr != nil {
			log.Panic(emailErr)
			c.JSON(http.StatusBadRequest, gin.H{"error": "error occurred while checking for this email"})
		}

		if emailCount > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Looks like this email already exists", "count": emailCount})
			return
		}

		passwordHash := MaskPassword(*user.Password)
		user.Password = &passwordHash

		// Validator
		validatorError := validate.Struct(&user)
		if validatorError != nil {
			c.JSON(
				http.StatusBadRequest,
				gin.H{
					"Message": "Error",
					"Data":    map[string]interface{}{"data": validatorError.Error()},
				},
			)
		}

		// Insert User
		newUser := models.User{
			ID:       user.ID,
			Name:     user.Name,
			Email:    user.Email,
			Password: user.Password,
			// Uncomment this
			// Age:      user.Age,
		}

		result, err := userCollection.InsertOne(ctx, newUser)

		if err != nil {
			c.JSON(
				http.StatusBadRequest,
				gin.H{
					"Message": "error",
					"Data":    map[string]interface{}{"data": validatorError.Error()},
				},
			)
			return
		}

		c.JSON(
			http.StatusCreated,
			gin.H{
				"Data": map[string]interface{}{"data": result},
			},
		)
	}
}
