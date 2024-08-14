package router

import (
	"sunoapi/controller"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func CORS() gin.HandlerFunc {
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowCredentials = true
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"*"}
	return cors.New(config)
}

func SetupRouter(r *gin.Engine) {
	r.Use(CORS())
	r.POST("/v1/chat/completions", controller.ChatCompletions)
	r.GET("/fetch/{task_id}", controller.FetchTask)
}
