package helpers

import (
	"context"
	"log"
	"os"
	"shive/database"
	"shive/models"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type JwtSignedDetails struct {
	Email     string
	Name      string
	Username  string
	Uid       string
	User_type string
	jwt.StandardClaims
}

var userCollection *mongo.Collection = database.OpenCollection(database.Client, "user")

var SECRET_KEY string = os.Getenv("SECRET_KEY")

func GenerateAllTokens(
	email string,
	name string,
	userName string,
	userType string,
	uid string) (
	signedToken string,
	signedRefreshToken string,
	err error,
) {

	claims := &JwtSignedDetails{
		Uid:       uid,
		Email:     email,
		Name:      name,
		Username:  userName,
		User_type: userType,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Local().Add(time.Hour * time.Duration(12)).Unix(),
		},
	}

	refreshClaims := &JwtSignedDetails{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Local().Add(time.Hour * time.Duration(100)).Unix(),
		},
	}

	token, err := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		claims,
	).SignedString(
		[]byte(SECRET_KEY),
	)
	refreshToken, refreshTokenErr := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString([]byte(SECRET_KEY))

	if err != nil {

		log.Panic(err)

		return
	}
	if refreshTokenErr != nil {
		log.Panic(refreshTokenErr)
		return
	}

	return token, refreshToken, err

}

func ValidateToken(signedToken string) (claims *JwtSignedDetails, msg string) {
	token, err := jwt.ParseWithClaims(
		signedToken,
		&JwtSignedDetails{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte(SECRET_KEY), nil
		},
	)

	if err != nil {
		msg = err.Error()
		return nil, msg
	}

	claims, ok := token.Claims.(*JwtSignedDetails)

	if !ok {
		msg = "This token is incorrect. Sorry"
		return nil, msg
	}

	if claims.ExpiresAt < time.Now().Local().Unix() {

		msg = "Ooooops looks like your token has expired"
		return nil, msg

	}

	return claims, msg
}

func UpdateTokens(signedToken string, signedRefreshedToken string, userId string) (*models.User, error) {
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	// Initialize the update document
	updateTokenDocs := bson.D{
		{
			Key:   "token",
			Value: signedToken,
		},
		{
			Key:   "refreshedToken",
			Value: signedRefreshedToken,
		},
		{
			Key:   "updated_at",
			Value: time.Now()},
	}

	// Define filter based on the userId
	filter := bson.M{"user_id": userId}
	returnDocument := options.After

	// Specify options for upsert and to return the updated document
	upsert := true
	opt := options.FindOneAndUpdateOptions{
		Upsert:         &upsert,
		ReturnDocument: &returnDocument, // Return the updated document
	}

	// Perform the update operation and get the updated document back
	var updatedUser models.User
	err := userCollection.FindOneAndUpdate(
		ctx,
		filter,
		bson.D{
			{
				Key:   "$set",
				Value: updateTokenDocs,
			},
		},
		&opt,
	).Decode(&updatedUser) // Decode the result into updatedUser

	if err != nil {
		log.Printf("Error updating token for user %s: %v", userId, err.Error())
		return nil, err
	}

	return &updatedUser, nil
}
