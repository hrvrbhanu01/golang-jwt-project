package controllers

import (
	"context"
	"fmt"
	helper "golang-jwt-project/helpers"
	"log"
	"net/http"
	"strconv"
	"time"

	database "golang-jwt-project/database"
	model "golang-jwt-project/models"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

var userCollection *mongo.Collection = database.OpenCollection(database.Client, "user")
var validate=validator.New()

func hashPassword(password string)string{
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err!=nil{
		log.Panic(err)
	}

	return string(bytes)
}

func VerifyPassword(userPassword string, providedPassword string)(bool, string) {   //check is bool here and msg is string,         //the second () is used for what we are going to return
	err := bcrypt.CompareHashAndPassword([]byte(providedPassword), []byte(userPassword))
	check := true
	msg := ""

	if err!=nil{
		msg=fmt.Sprintf("email or password is incorrect! Error: %v", err)
		check=false

	}
	return check, msg
}

func Signup()gin.HandlerFunc{
	return func(c *gin.Context){
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var user model.User

		if err:=c.BindJSON(&user); err !=nil{
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		validationErr:=validate.Struct(user)
		if validationErr!=nil{
			c.JSON(http.StatusBadRequest, gin.H{"erroerrr":validationErr.Error()})
			return
		}	
		emailCount, err := userCollection.CountDocuments(ctx, bson.M{"email": user.Email})
        if err != nil {
            log.Printf("Error checking email: %v", err)
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error occurred"})
            return
        }

		phoneCount, err := userCollection.CountDocuments(ctx, bson.M{"phone": user.Phone})
        if err != nil {
            log.Printf("Error checking phone: %v", err)
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error occurred"})
            return
        }


		
		if emailCount > 0 || phoneCount > 0 {
            if emailCount > 0 && phoneCount == 0 {
                c.JSON(http.StatusInternalServerError, gin.H{"error": "Email address already in use"})
            } else if emailCount == 0 && phoneCount > 0 {
                c.JSON(http.StatusInternalServerError, gin.H{"error": "Phone number already registered"})
            } else {
                c.JSON(http.StatusInternalServerError, gin.H{"error": "Email address or phone number already exists"})
            }
            return
        }
		password := hashPassword(*user.Password)   // Assuming a hashPassword function exists
		user.Password = &password  // Assign hashed password to the user object
		
		user.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))

		user.ID = primitive.NewObjectID()
		userID := user.ID.Hex()
		user.User_id = &userID
		token, refreshToken, _ := helper.GenerateAllTokens(*user.Email, *user.First_name, *user.Last_name, *user.User_type, *user.User_id)
		user.Token = &token
		user.Refresh_token = &refreshToken
		// to insert it into the DB
		resultInsertionNumber, insertErr := userCollection.InsertOne(ctx, user)
		if insertErr != nil {
			msg := fmt.Sprintf("User item was not created! Error: %v", insertErr)
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}
		defer cancel()
		c.JSON(http.StatusOK, resultInsertionNumber)
		

	}
}

func Login() gin.HandlerFunc{
	return func(c *gin.Context){
		var ctx, cancel=context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()   // Ensure cancellation happens even on successful login

		var user model.User
		var foundUser model.User

		if err := c.BindJSON(&user); err != nil{
			c.JSON(http.StatusBadRequest, gin.H{"error":err.Error()})
			return
		}

		err:=userCollection.FindOne(ctx, bson.M{"email":user.Email}).Decode(&foundUser)
		
		if err != nil{
			c.JSON(http.StatusInternalServerError, gin.H{"error": "email or password is incorrect!"})
			return
		}
		passwordIsValid, msg := VerifyPassword(*user.Password, *foundUser.Password)

		if !passwordIsValid{
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}

		if foundUser.Email==nil{
			c.JSON(http.StatusInternalServerError, gin.H{"error":"user not found!"})

		}
		token, refreshToken, _ := helper.GenerateAllTokens(*foundUser.Email, *foundUser.First_name, *foundUser.Last_name, *foundUser.User_type, *foundUser.User_id)
		helper.UpdateAllTokens(token, refreshToken, *foundUser.User_id)
		
		err = userCollection.FindOne(ctx, bson.M{"user_id":foundUser.User_id}).Decode(&foundUser)

		if err != nil{
			c.JSON(http.StatusInternalServerError, gin.H{"error":err.Error()})
			return
		}
		c.JSON(http.StatusOK, foundUser)
	}
}

func GetUsers() gin.HandlerFunc {
    return func(c *gin.Context) {
        if err := helper.CheckUserType(c, "ADMIN"); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }

        var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
        defer cancel() // Ensure the context is canceled

        recordPerPage, err := strconv.Atoi(c.Query("recordPerPage"))
        if err != nil || recordPerPage < 1 {
            recordPerPage = 10
        }
        page, err1 := strconv.Atoi(c.Query("page"))
        if err1 != nil || page < 1 {
            page = 1
        }

        startIndex := (page - 1) * recordPerPage

        // MongoDB aggregation stages
        matchStage := bson.D{{Key: "$match", Value: bson.D{}}}
        groupStage := bson.D{{Key: "$group", Value: bson.D{
            {Key: "_id", Value: bson.D{{Key: "_id", Value: "null"}}},
            {Key: "total_count", Value: bson.D{{Key: "$sum", Value: 1}}},
            {Key: "data", Value: bson.D{{Key: "$push", Value: "$$ROOT"}}},
        }}}
        projectStage := bson.D{{Key: "$project", Value: bson.D{
            {Key: "_id", Value: 0},
            {Key: "total_count", Value: 1},
            {Key: "user_items", Value: bson.D{{Key: "$slice", Value: []interface{}{"$data", startIndex, recordPerPage}}}},
        }}}

        cursor, err := userCollection.Aggregate(ctx, mongo.Pipeline{
            matchStage, groupStage, projectStage,
        })
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "error occurred while listing user items!"})
            return
        }
        defer cursor.Close(ctx)

        var allUsers []bson.M
        if err = cursor.All(ctx, &allUsers); err != nil {
            log.Fatal(err)
        }

        c.JSON(http.StatusOK, allUsers) // Send all users to the frontend or postman
    }
}

func GetUser() gin.HandlerFunc{
	return func(c *gin.Context){
		userId := c.Param("user_id")

		if err := helper.MatchUserTypeToUid(c, userId); err !=nil{
			c.JSON(http.StatusBadRequest, gin.H{"error":err.Error()})
			return
		}
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

		var user model.User
		err:=userCollection.FindOne(ctx, bson.M{"user_id":userId}).Decode(&user)
		defer cancel()
		if err!=nil{
			c.JSON(http.StatusInternalServerError, gin.H{"error":err.Error()})
			return
		}
		c.JSON(http.StatusOK, user)

	}
}