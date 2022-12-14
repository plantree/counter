package main

import (
	"crypto/md5"
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

func incrMethodCalls(method string) error {
	_, err := G_db.Incr("call@" + method)
	if err != nil {
		G_logger.Errorf("incr method calls failed. err: %v", err)
	}
	return err
}

func constructNamespace(namespace string) string {
	return fmt.Sprintf("namespace@%s", namespace)
}

func constructKey(namespace, key string) string {
	return fmt.Sprintf("key@%s@%s", namespace, key)
}

func isNamespaceValid(namespace string, c *gin.Context) bool {
	namespace = constructNamespace(namespace)
	_, err := G_db.Get(namespace)
	if err != nil {
		G_logger.Warn(err)
		c.JSON(http.StatusBadRequest, "invalid namespace")
		return false
	}
	return true
}

// encrypt the password
func generateScrect(namespace string, secret string) (string, error) {
	h := md5.New()
	_, err := h.Write([]byte(namespace + secret))
	if err != nil {
		return "", err
	}
	return string(h.Sum(nil)), nil
}

func checkNamespace(namespace string, c *gin.Context) bool {
	if namespace == "" || strings.Contains(namespace, "@") {
		c.JSON(http.StatusBadRequest, "need namespace without @")
		return false
	}
	return true
}

func checkNamespaceAndKey(namespace, key string, c *gin.Context) bool {
	if namespace == "" || strings.Contains(namespace, "@") || key == "" {
		c.JSON(http.StatusBadRequest, "need namespace without @ and key")
		return false
	}
	return true
}

func checkNamespaceSecretKey(namespace, secret, key string, c *gin.Context) bool {
	if namespace == "" || strings.Contains(namespace, "@") ||
		secret == "" || key == "" {
		c.JSON(http.StatusBadRequest, "need namespace without @, secret and key")
		return false
	}
	return true
}

func checkNamespaceSecretKeyValue(namespace, secret, key, value string, c *gin.Context) bool {
	if namespace == "" || strings.Contains(namespace, "@") ||
		secret == "" || key == "" || value == "" || !isInt(value) {
		c.JSON(http.StatusBadRequest, "need namespace without @, secret, key and valid value (integer)")
		return false
	}
	return true
}

func checkAuthentication(namespace, secret string) bool {
	namespace = constructNamespace(namespace)
	result, err := G_db.Get(namespace)
	if err != nil {
		G_logger.Warn(err)
		return false
	}
	compareScript, err := generateScrect(namespace, secret)
	if err != nil {
		G_logger.Warn(err)
		return false
	}
	if result.value.(string) != compareScript {
		errMsg := fmt.Errorf("secret is invalid")
		G_logger.Warn(errMsg)
		return false
	}
	return true
}

func GetPv(c *gin.Context) {
	incrMethodCalls("get_pv")

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
		newKeys, err := G_db.GetPrefixMatchKeys(fmt.Sprintf("key@%s@*", namespace))
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
	incrMethodCalls("create_pv")

	namespace := c.Query("namespace")
	if ok := checkNamespace(namespace, c); !ok {
		return
	}
	secret := c.Query("secret")
	if secret == "" {
		secret = namespace
	}
	namespace = constructNamespace(namespace)
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
	secret, err := generateScrect(namespace, secret)
	if err != nil {
		G_logger.Warn(err)
		errMsg := ErrorMessage{
			Code:   5001,
			ErrMsg: "internal error",
		}
		c.JSON(http.StatusInternalServerError, errMsg)
		return
	}
	err = G_db.Set(namespace, secret, false)
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
	incrMethodCalls("increment_pv")

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
		return
	}
	errMsg := ErrorMessage{
		Code:   0,
		ErrMsg: "incr key successfully",
	}
	errMsg.Data = append(errMsg.Data, Data{Key: result.key, Value: result.value})
	c.JSON(http.StatusOK, errMsg)
}

func ResetPv(c *gin.Context) {
	incrMethodCalls("reset_pv")

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
		return
	}
	errMsg := ErrorMessage{
		Code:   0,
		ErrMsg: "reset key successfully",
	}
	c.JSON(http.StatusOK, errMsg)
}

func DeletePv(c *gin.Context) {
	incrMethodCalls("delete_pv")

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
		return
	}
	errMsg := ErrorMessage{
		Code:   0,
		ErrMsg: fmt.Sprintf("delete %d keys successfully", cnt),
	}
	c.JSON(http.StatusOK, errMsg)
}

func CountNamespaces(c *gin.Context) {
	allKeys, err := G_db.GetPrefixMatchKeys("namespace@*")
	if err != nil {
		errMsg := ErrorMessage{
			Code:   5001,
			ErrMsg: fmt.Sprintf("internal error. err: %v", err),
		}
		c.JSON(http.StatusInternalServerError, errMsg)
		return
	}
	errMsg := ErrorMessage{
		Code:   0,
		ErrMsg: "count namespaces successfully",
		Data: []Data{
			{Key: "namespace count", Value: len(allKeys)},
		},
	}
	c.JSON(http.StatusOK, errMsg)
}

func CountKeys(c *gin.Context) {
	allKeys, err := G_db.GetPrefixMatchKeys("key@*")
	if err != nil {
		errMsg := ErrorMessage{
			Code:   5001,
			ErrMsg: fmt.Sprintf("internal error. err: %v", err),
		}
		c.JSON(http.StatusInternalServerError, errMsg)
		return
	}
	errMsg := ErrorMessage{
		Code:   0,
		ErrMsg: "count keys successfully",
		Data: []Data{
			{Key: "keys count", Value: len(allKeys)},
		},
	}
	c.JSON(http.StatusOK, errMsg)
}

func CountRequests(c *gin.Context) {
	allKeys, err := G_db.GetPrefixMatchKeys("call@*")
	if err != nil {
		errMsg := ErrorMessage{
			Code:   5001,
			ErrMsg: fmt.Sprintf("internal error. err: %v", err),
		}
		c.JSON(http.StatusInternalServerError, errMsg)
		return
	}
	requestCount := 0
	for _, key := range allKeys {
		result, err := G_db.Get(key)
		if err != nil {
			G_logger.Errorf("get key %s failed. err: %v", key, err)
			continue
		}
		if value, err := strconv.Atoi(result.value.(string)); err == nil {
			requestCount += value
		}
	}
	errMsg := ErrorMessage{
		Code:   0,
		ErrMsg: "count requests successfully",
		Data: []Data{
			{Key: "requests count", Value: requestCount},
		},
	}
	c.JSON(http.StatusOK, errMsg)
}

func AddRouters(r *gin.Engine, logger *logrus.Logger) {
	G_db = NewRedisClient(DEFAULT_REDIS_URL, logger)
	if G_db == nil {
		panic("get redis client failed")
	}
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

	// statistics API
	r.GET("/pv/statistics/count-namespaces", CountNamespaces)

	r.GET("/pv/statistics/count-keys", CountKeys)

	r.GET("/pv/statistics/count-requests", CountRequests)
}
