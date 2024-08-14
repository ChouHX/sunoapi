package main

import (
	"fmt"
	"net/http"
	"sunoapi/common"
	"sunoapi/middleware"
	"sunoapi/router"

	"github.com/gin-gonic/gin"
)

func Init() *gin.Engine {
	common.InitTemplate()
	common.HTTPClient = &http.Client{}

	server := gin.New()
	server.Use(middleware.RequestId())
	router.SetupRouter(server)
	return server
}

func main() {
	server := Init()
	common.LogSuccess(fmt.Sprintf("BaseUrl:%s", common.BaseUrl))
	common.LogSuccess(fmt.Sprintf("Start: 0.0.0.0:" + common.Port))

	err := server.Run(":" + common.Port)
	if err != nil {
		common.LogError("failed to start HTTP server: " + err.Error())
	}
}
