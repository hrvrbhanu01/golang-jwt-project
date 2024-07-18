package middleware

import(
	"fmt"
	"net/http"
	helper "golang-jwt-project/helpers"
	"github.com/gin-gonic/gin"
)

func Authenticate() gin.HandlerFunc{  //before calling the userRoutes we are authenticating it
	return func(c *gin.Context){
		clientToken := c.Request.Header.Get("token")
		if clientToken == ""{
			c.JSON(http.StatusInternalServerError, gin.H{"error":fmt.Sprintf("No Authorization header provided!")})
			c.Abort()
			return
		}

		claims, err := helper.ValidateToken(clientToken)
		
	}
}