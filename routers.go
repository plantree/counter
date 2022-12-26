package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Data struct {
	Key   string      `json:"key"`
	Value interface{} `json:"value"`
}

type ErrorMessage struct {
	Code   int32  `json:"code"`
	ErrMsg string `json:"err_msg"`
	Data   []Data `json:"data"`
}

var G_db *DB
var G_logger *logrus.Logger

func isInt(s string) bool {
	if _, err := strconv.Atoi(s); err != nil {
		return false
	}
	return true
}

func isNamespaceValid(namespace string, c *gin.Context) bool {
	_, err := G_db.Get(namespace)
	if err != nil {
		G_logger.Warn(err)
		c.JSON(http.StatusBadRequest, "invalid namespace")
		return false
	}
	return true
}

func constructKey(namespace, key string) string {
	return fmt.Sprintf("%s@%s", namespace, key)
}

func checkNamespace(namespace string, c *gin.Context) bool {
	if namespace == "" || strings.Contains(namespace, "@") {
		c.JSON(http.StatusBadRequest, "need namespace (namespace should not contain @)")
		return false
	}
	return true
}

func checkNamespaceAndKey(namespace, key string, c *gin.Context) bool {
	if namespace == "" || strings.Contains(namespace, "@") || key == "" {
		c.JSON(http.StatusBadRequest, "need namespace and key (namespace should not contain @)")
		return false
	}
	return true
}

func checkNamespaceSecretKey(namespace, secret, key string, c *gin.Context) bool {
	if namespace == "" || strings.Contains(namespace, "@") ||
		secret == "" || key == "" {
		c.JSON(http.StatusBadRequest, "need namespace, secret and key (namespace should not contain @)")
		return false
	}
	return true
}

func checkNamespaceSecretKeyValue(namespace, secret, key, value string, c *gin.Context) bool {
	if namespace == "" || strings.Contains(namespace, "@") ||
		secret == "" || key == "" || value == "" || !isInt(value) {
		c.JSON(http.StatusBadRequest, "need namespace, secret, key and valid value (namespace should not contain @)")
		return false
	}
	return true
}

func checkAuthentication(namespace, secret string) bool {
	result, err := G_db.Get(namespace)
	if err != nil {
		G_logger.Warn(err)
		return false
	}
	if result.value.(string) != secret {
		errMsg := fmt.Errorf("secret is invalid")
		G_logger.Warn(errMsg)
		return false
	}
	return true
}

func GetPv(c *gin.Context) {
	namespace := c.Query("namespace")
	if ok := checkNamespace(namespace, c); !ok {
		return
	}
	if ok := isNamespaceValid(namespace, c); !ok {
		return
	}

	key := c.Query("key")
	// get all keys under namespace
	if key == "" {
		newKeys, err := G_db.GetPrefixMatchKeys(namespace + "@*")
		if err != nil {
			G_logger.Warn(err)
			errMsg := ErrorMessage{
				Code:   5001,
				ErrMsg: "internal error",
			}
			c.JSON(http.StatusInternalServerError, errMsg)
			return
		}
		if len(newKeys) == 0 {
			errMsg := ErrorMessage{
				Code:   4001,
				ErrMsg: "this namespace or key doesn't exist",
			}
			c.JSON(http.StatusBadRequest, errMsg)
			return
		}
		// batch get
		results, err := G_db.BatchGet(newKeys...)
		if err != nil {
			G_logger.Warn(err)
			errMsg := ErrorMessage{
				Code:   5001,
				ErrMsg: "internal error",
			}
			c.JSON(http.StatusInternalServerError, errMsg)
			return
		}
		errMsg := ErrorMessage{
			Code:   0,
			ErrMsg: "successfully",
		}
		for _, item := range results {
			errMsg.Data = append(errMsg.Data, Data{Key: item.key, Value: item.value})
		}
		c.JSON(http.StatusOK, errMsg)
		return
	}
	// specific key
	newKey := constructKey(namespace, key)
	result, err := G_db.Get(newKey)
	if err != nil {
		G_logger.Warn(err)
		if strings.Contains(err.Error(), "does not exist") {
			errMsg := ErrorMessage{
				Code:   4001,
				ErrMsg: "this namespace or key doesn't exist",
			}
			c.JSON(http.StatusBadRequest, errMsg)
			return
		}
		errMsg := ErrorMessage{
			Code:   5001,
			ErrMsg: "internal error",
		}
		c.JSON(http.StatusInternalServerError, errMsg)
		return
	}
	errMsg := ErrorMessage{
		Code:   0,
		ErrMsg: "get key successfully",
	}
	errMsg.Data = append(errMsg.Data, Data{Key: result.key, Value: result.value})
	c.JSON(http.StatusOK, errMsg)
}

func CreatePv(c *gin.Context) {
	namespace := c.Query("namespace")
	if ok := checkNamespace(namespace, c); !ok {
		return
	}
	secret := c.Query("secret")
	if secret == "" {
		secret = namespace
	}
	result, _ := G_db.Get(namespace)
	if result != nil {
		G_logger.Warn(fmt.Sprintf("namespace[%s] exists", namespace))
		errMsg := ErrorMessage{
			Code:   4001,
			ErrMsg: "this namespace exists",
		}
		c.JSON(http.StatusBadRequest, errMsg)
		return
	}

	err := G_db.Set(namespace, secret, false)
	if err != nil {
		G_logger.Warn(err)
		errMsg := ErrorMessage{
			Code:   5001,
			ErrMsg: "internal error",
		}
		c.JSON(http.StatusInternalServerError, errMsg)
		return
	}
	errMsg := ErrorMessage{
		Code:   0,
		ErrMsg: "create namespace successfully",
	}
	errMsg.Data = append(errMsg.Data, Data{Key: namespace, Value: secret})
	c.JSON(http.StatusOK, errMsg)
}

func IncrementPv(c *gin.Context) {
	namespace := c.Query("namespace")
	key := c.Query("key")
	if ok := checkNamespaceAndKey(namespace, key, c); !ok {
		return
	}
	if ok := isNamespaceValid(namespace, c); !ok {
		return
	}

	newKey := constructKey(namespace, key)
	result, err := G_db.Incr(newKey)
	if err != nil {
		G_logger.Warn(err)
		errMsg := ErrorMessage{
			Code:   5001,
			ErrMsg: "internal error",
		}
		c.JSON(http.StatusInternalServerError, errMsg)
	}
	errMsg := ErrorMessage{
		Code:   0,
		ErrMsg: "incr key successfully",
	}
	errMsg.Data = append(errMsg.Data, Data{Key: result.key, Value: result.value})
	c.JSON(http.StatusOK, errMsg)
}

func ResetPv(c *gin.Context) {
	namespace := c.Query("namespace")
	secret := c.Query("secret")
	key := c.Query("key")
	value := c.Query("value")
	if ok := checkNamespaceSecretKeyValue(namespace, secret, key, value, c); !ok {
		return
	}

	if ok := checkAuthentication(namespace, secret); !ok {
		errMsg := ErrorMessage{
			Code:   4002,
			ErrMsg: "authentication failed",
		}
		c.JSON(http.StatusBadRequest, errMsg)
		return
	}

	newKey := constructKey(namespace, key)
	if err := G_db.Set(newKey, value, true); err != nil {
		errMsg := ErrorMessage{
			Code:   5001,
			ErrMsg: "internal error",
		}
		c.JSON(http.StatusInternalServerError, errMsg)
	}
	errMsg := ErrorMessage{
		Code:   0,
		ErrMsg: "reset key successfully",
	}
	c.JSON(http.StatusOK, errMsg)
}

func DeletePv(c *gin.Context) {
	namespace := c.Query("namespace")
	secret := c.Query("secret")
	key := c.Query("key")
	if ok := checkNamespaceSecretKey(namespace, secret, key, c); !ok {
		return
	}

	if ok := checkAuthentication(namespace, secret); !ok {
		errMsg := ErrorMessage{
			Code:   4002,
			ErrMsg: "authentication failed",
		}
		c.JSON(http.StatusBadRequest, errMsg)
		return
	}

	newKey := constructKey(namespace, key)
	cnt, err := G_db.Delete(newKey)
	if err != nil {
		errMsg := ErrorMessage{
			Code:   5001,
			ErrMsg: "internal error",
		}
		c.JSON(http.StatusInternalServerError, errMsg)
	}
	errMsg := ErrorMessage{
		Code:   0,
		ErrMsg: fmt.Sprintf("delete %d keys successfully", cnt),
	}
	c.JSON(http.StatusOK, errMsg)
}

func AddRouters(r *gin.Engine, logger *logrus.Logger) {
	G_db = NewRedisClient(DEFAULT_REDIS_URL, logger)
	G_logger = logger

	// server status
	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	// api for PV
	r.GET("/pv/get", GetPv)

	r.POST("/pv/create", CreatePv)

	r.POST("/pv/increment", IncrementPv)

	r.POST("/pv/reset", ResetPv)

	r.POST("/pv/delete", DeletePv)
}
