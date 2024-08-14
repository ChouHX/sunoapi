package common

import (
	"fmt"
	"sunoapi/lib/ginplus"

	"github.com/gin-gonic/gin"
)

const (
	ColorReset = "\033[0m"
	ColorRed   = "\033[31m"
	ColorGreen = "\033[32m"
)

func LogSuccess(log string) {
	fmt.Printf(ColorGreen+"[ SunoAPI ] %s \n"+ColorReset, log)
}

func LogError(log string) {
	fmt.Printf(ColorRed+"[ SunoAPI ] %s \n"+ColorReset, log)
}

type RelayError struct {
	StatusCode int
	Code       string
	Err        error
	LocalErr   bool
}

const (
	ErrCodeInvalidRequest = "invalid_request"
	ErrCodeInternalError  = "internal_error"
)

func ReturnErr(c *gin.Context, err error, code string, statusCode int) {
	c.JSON(statusCode, ginplus.BuildApiReturn(code, err.Error(), nil))
}

func ReturnRelayErr(c *gin.Context, relayErr *RelayError) {
	if relayErr.Err == nil {
		relayErr.Err = fmt.Errorf("unknown error")
	}
	c.JSON(relayErr.StatusCode, ginplus.BuildApiReturn(relayErr.Code, relayErr.Err.Error(), nil))
}

func WrapperErr(err error, code string, statusCode int) *RelayError {
	return &RelayError{
		StatusCode: statusCode,
		Code:       code,
		Err:        err,
	}
}
