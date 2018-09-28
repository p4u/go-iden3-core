package endpoint

import (
	"github.com/gin-gonic/gin"
	"github.com/iden3/go-iden3/cmd/cauth/config"
)

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Add("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Add("Origin", "*")
		c.Writer.Header().Add("X-Requested-With", "*")
		c.Next()
	}
}

func Serve() {

	r := gin.Default()
	r.Use(corsMiddleware())
	r.POST("/auth", handleAuth)
	r.GET("/ws/:id", handleWs)
	r.Run(config.C.Server.ServiceApi)
}
