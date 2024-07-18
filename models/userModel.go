package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct{
	ID				primitive.ObjectID		`bson:"id"`
	First_name		*string					`json:"first_name" validate:"required,min=2,max=100"` //this is for database
	Last_name		*string					`json:"last_name" validate:"required,min=2,max=100"`
	Password		*string					`json:"password" validate:"required,min=6"`
	Email			*string					`json:"email" validate:"email,required"` //in validate : email is a validation type itself to check if the provided value is email type ot not!
	Phone			*string					`json:"phone" validate:"required"`
	Token			*string					`json:"token"`
	User_type		*string					`json:"user_type" validate:"required,eq=ADMIN|eq=USER"` //similar as enum in Javascript 
	Refresh_token	*string					`json:"refresh_token"`
	Created_at		time.Time				`json:"created_at"`
	Updated_at		time.Time				`json:"updated_at"`
	User_id			*string    		 	  	`json:"user_id"`
}