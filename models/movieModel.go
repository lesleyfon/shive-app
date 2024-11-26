package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Movie struct {
	Id    primitive.ObjectID `bson:"id"`
	Name  *string            `json:"name" validate:"required"`
	Topic *string            `json:"topic" validate:"required"`

	Movie_id  string `json:"movie_id"`
	Movie_URL string `json:"movie_url"`

	Genre_id string `json:"genre_id"`

	Created_at time.Time `json:"created_at"`
	Updated_at time.Time `json:"updated_at"`
}
