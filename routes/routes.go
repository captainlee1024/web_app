// Package routes provides ...
package routes

import (
	"net/http"

	"github.com/captainlee1024/web_app/logger"
	"github.com/gin-gonic/gin"
)

// Setup 注册路由
func Setup() *gin.Engine {
	r := gin.New()
	r.Use(logger.GinLogger(), logger.GinRecovery(true))

	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"msg": "ok",
		})
	})
	return r
}
