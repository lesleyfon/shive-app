package middleware

import (
	"net/http"
	"shive/helpers"

	"github.com/gin-gonic/gin"
)

// The `Authenticate` checks for a valid token in the request header and validates it
// before setting user information in the context.
func Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientToken := c.Request.Header.Get("token")

		if clientToken == "" {
			c.JSON(
				http.StatusBadRequest,
				gin.H{
					"error": "No Authorization header provided",
				},
			)
			c.Abort()
			return
		}
		claims, err := helpers.ValidateToken(clientToken)

		if err != "" {
			c.JSON(
				http.StatusBadRequest,
				gin.H{
					"error": err,
				},
			)
			c.Abort()
			return
		}

		c.Set("email", claims.Email)
		c.Set("username", claims.Username)
		c.Set("name", claims.Name)
		c.Set("uid", claims.Uid)
		c.Set("user_type", claims.User_type)
		c.Next()

	}
}
