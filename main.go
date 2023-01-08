/*
 * Counter API
 *
 * This is a simple API for [counter](https://github.com/plantree/counter)
 *
 * API version: 1.0.0
 * Contact: eric.wangpy@outlook.com
 */
package main

import (
	"os"

	"github.com/gin-gonic/gin"
)

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func Init() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)

	r := gin.Default()
	logger := NewLogger()
	r.Use(LoggerMiddleware(logger))
	r.Use(CORSMiddleware())

	AddRouters(r, logger)

	return r
}

func main() {
	r := Init()

	if environment := os.Getenv("GIN_MODE"); environment == "release" {
		r.Run(":80")
	} else {
		r.Run(SERVER_PORT)
	}
}
