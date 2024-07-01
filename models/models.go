// models.go
// Author: Bipin Kumar Ojha (Freelancer)

package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type User struct {
	ID       primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Username string             `json:"username" bson:"username"`
	Password string             `json:"password" bson:"password"`
}

type Task struct {
	ID          primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	UserID      primitive.ObjectID `json:"userId" bson:"userId"`
	Title       string             `json:"title" bson:"title"`
	Description string             `json:"description" bson:"description"`
	AllottedTo  string             `json:"allotted_to" bson:"allotted_to"`
	DoneBy      string             `json:"done_by" bson:"done_by"`
	Status      string             `json:"status" bson:"status"`
	StartDate   primitive.DateTime `json:"start_time" bson:"start_time"`
	EndDate     primitive.DateTime `json:"end_time" bson:"end_time"`
}
