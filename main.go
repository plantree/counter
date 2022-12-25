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
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	logger := NewLogger()
	r.Use(LoggerMiddleware(logger))

	AddRouters(r, logger)

	r.Run(SERVER_PORT)
}
