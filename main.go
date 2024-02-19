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
	"time"

	"github.com/gin-gonic/gin"
	"github.com/juju/ratelimit"
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

func RateLimitMiddleware(fillInterval time.Duration, cap, quantum int64) gin.HandlerFunc {
	bucket := ratelimit.NewBucketWithQuantum(fillInterval, cap, quantum)
	return func(c *gin.Context) {
		if bucket.TakeAvailable(1) < 1 {
			c.AbortWithStatus(429)
			return
		}
		c.Next()
	}
}

func Init() *gin.Engine {
	// gin.SetMode(gin.ReleaseMode)

	r := gin.Default()
	logger := NewLogger()
	r.Use(LoggerMiddleware(logger))
	r.Use(CORSMiddleware())
	r.Use(RateLimitMiddleware(1*time.Second, 100, 50))

	AddRouters(r, logger)

	return r
}

func main() {
	r := Init()

	if environment := os.Getenv("GIN_MODE"); environment == "release" {
		r.Run(":8000")
	} else {
		r.Run(SERVER_PORT)
	}
}
