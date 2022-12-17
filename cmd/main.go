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

	"github.com/plantree/counter/config"
	"github.com/plantree/counter/middleware"
)

func main() {
	r := gin.Default()
	logger := middleware.NewLogger()
	r.Use(middleware.LoggerMiddleware(logger))
	middleware.AddRouters(r, logger)

	r.Run(config.SERVER_PORT)
}
