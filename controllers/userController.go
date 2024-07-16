package controllers

import (
	"context"
	"fmt"
	"golang-jwt-project/bcrypt"
	"golang-jwt-project/helpers"
	helper "golang-jwt-project/helpers"
	"golang-jwt-project/models"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/hrvrbhanu01/golang-jwt-project/database"
	"github.com/hrvrbhanu01/golang-jwt-project/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/bson/primitive"
	//"github.com/hrvrbhanu01/golang-jwt-project/database"
)

var userCollection *mongo.Collection = database.OpenCollection(database.Client, "user")
var validate=validator.New()

func HashPassword()

func VerifyPassword()

func SignUp()gin.HandlerFunc{
	return func(c *gin.Context){
		var ctx, cancel=context.WithTimeout(context.Background(), 100*time.Second)
		var user models.User

		if err:=c.BindJSON(&user); err !=nil{
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		validationErr:=validate.Struct(user)
		if validationErr!=nil{
			c.JSON(http.StatusBadRequest, gin.H{"error":validationErr.Error()})
			return
		}	
		count, err := userCollection.CountDocuments(ctx, bson.M{"email":user.Email})
		defer cancel()
		if err!=nil{
			log.Panic(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error":"error occurred while checking for the email"})
		}

		count, err = userCollection.CountDocuments(ctx, bson.M{"phone":user.Phone})
		defer cancel()
		if err!=nil{
			log.Panic(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error":"error occurred while checking for the phone number!"})

		}
		if count >0{
			c.JSON(http.StatusInternalServerError, gin.H{"error":"this email or phone number already exists!"})
		}

		user.Created_at, _=time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.Updated_at, _=time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.ID=primitive.NewObjectID()
		user.User_id=user.ID.Hex()
		token, refreshToken,_ :=helper.GenerateAllTokens(*user.Email, *user.First_name, *user.Last_name, *user.User_type, *&user.User_id)
		user.Token=&token
		user.Refresh_token=&refreshToken
		//to insert it into the DB
		resultInsertionNumber, insertErr:=userCollection.InsertOne(ctx, user)
		if insertErr!=nil{
			msg:=fmt.Sprintf("Useritem was not created!")
			c.JSON(http.StatusInternalServerError, gin.H{"error":msg})
			return
		}
		defer cancel()
		c.JSON(http.StatusOK, resultInsertionNumber)

	}
}

func Login()

func GetUsers()

func GetUser() gin.HandlerFunc{
	return func(c *gin.Context){
		userId := c.Param("user_id")

		if err:=helper.MatchUserTypeToUid(c, userId); err !=nil{
			c.JSON(http.StatusBadRequest, gin.H{"error":err.Error()})
			return
		}
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

		var user models.User
		err:=user.Collection.FindOne(ctx, bson.M{"user_id":userId}).Decode(&user)
		defer cancel()
		if err!=nil{
			c.JSON(http.StatusInternalServerError, gin.H{"error":err.Error()})
			return
		}
		c.JSON(http.StatusOK, user)

	}
}