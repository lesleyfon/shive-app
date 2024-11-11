package helpers

import (
	"errors"

	"github.com/gin-gonic/gin"
)

// VerifyUserType checks if the user type in the context matches the specified  role and returns an error if they do not match.
func VerifyUserType(c *gin.Context, role string) (err error) {
	userType := c.GetString("user_type")
	err = nil

	if userType != role {
		err = errors.New("You Are Not Authorised to Access This!")
		return err
	}
	return err
}

func MatchToUid(c *gin.Context, userId string) (err error) {
	userType := c.GetString("user_type")
	uid := c.GetString("uid")

	err = nil

	if userId != uid {
		err = errors.New("You are not authorised to access this!Unauthorized to access this resource")
		return err

	}

	// This function checks if the user type in the context matches the specified role.
	err = VerifyUserType(c, userType)

	return err
}
