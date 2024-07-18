package routes

import (
	controller "golang-jwt-project/controllers"

	"github.com/gin-gonic/gin"
)

func authRoutes(incomingRoutes *gin.Engine){  //we are not authenticating these routes because these are public routes so anybody should access them
	incomingRoutes.POST("users/signup", controller.Signup())
	incomingRoutes.POST("users/login", controller.Login())
}