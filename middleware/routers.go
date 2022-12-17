package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type ErrorMessage struct {
	Code   int32  `json:"code"`
	ErrMsg string `json:"err_msg"`
}

func GetPv(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"pv": 100,
	})
}

func CreatePv(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"pv": 100,
	})
}

func IncrementPv(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"pv": 100,
	})
}

func ResetPv(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"pv": 100,
	})
}

func DeletePv(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"pv": 100,
	})
}

func AddRouters(r *gin.Engine, logger *logrus.Logger) {
	// server status
	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	// api for PV
	r.GET("/pv/status", GetPv)

	r.POST("/pv/create", CreatePv)

	r.POST("/pv/increment", IncrementPv)

	r.POST("/pv/reset", ResetPv)

	r.POST("/pv/delete", DeletePv)

}
