package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"sunoapi/common"

	"github.com/gin-gonic/gin"
)

func SecretAuth() func(c *gin.Context) {
	return func(c *gin.Context) {
		if common.SecretToken == "" {
			return
		}
		accessToken := c.Request.Header.Get("Authorization")
		accessToken = strings.TrimLeft(accessToken, "Bearer ")
		if accessToken == common.SecretToken {
			c.Next()
		} else {
			common.ReturnErr(c, fmt.Errorf("unauthorized secret token"), common.ErrCodeInvalidRequest, http.StatusUnauthorized)
			c.Abort()
			return
		}
	}
}
