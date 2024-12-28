package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Review struct {
	Id          primitive.ObjectID `bson:"id"`
	Review      *string            `json:"review"`
	Review_id   string             `json:"review_id"`
	Movie_id    string             `json:"movie_id"`
	Reviewer_id string             `json:"reviewer_id"`
	Created_at  time.Time          `json:"created_at"`
	Updated_at  time.Time          `json:"updated_at"`
}
