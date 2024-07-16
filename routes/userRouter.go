package routes

import(
	controller "golang-jwt-project/controllers"
	"golang-jwt-project/middleware"
	"github.com/gin-gonic/gin"
)

func userRoutes(incomingRoutes *gin.Engine){
	incomingRoutes.Use(middleware.Authenticate())
	incomingRoutes.GET("/users", controller.GetUsers())
	incomingRoutes.GET("/users/:user_id", controller.GetUser())
}