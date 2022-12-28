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

func Init() *gin.Engine {
	r := gin.Default()
	logger := NewLogger()
	r.Use(LoggerMiddleware(logger))

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
