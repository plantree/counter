package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func MockRouters() *gin.Engine {
	r := gin.Default()
	logger := NewLogger()
	r.Use(LoggerMiddleware(logger))

	AddRouters(r, logger)
	return r
}

func TestPingRoute(t *testing.T) {
	router := MockRouters()
	defer CleanLog()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ping", nil)
	router.ServeHTTP(w, req)

	if w.Code != 200 || w.Body.String() != "pong" {
		t.Fail()
	}
}

func TestCreatePv(t *testing.T) {
	router := MockRouters()
	defer CleanLog()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/pv/create", nil)
	router.ServeHTTP(w, req)

	if w.Code != 400 || !strings.Contains(w.Body.String(), "need namespace") {
		fmt.Println(w.Body.String())
		t.Fail()
	}

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/pv/create?namespace=test", nil)
	router.ServeHTTP(w, req)

	if w.Code != 200 || !strings.Contains(w.Body.String(), "successfully") {
		fmt.Println(w.Body.String())
		t.Fail()
	}
}

func TestIncr(t *testing.T) {
	router := MockRouters()
	defer CleanLog()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/pv/increment", nil)
	router.ServeHTTP(w, req)

	if w.Code != 400 || !strings.Contains(w.Body.String(), "need") {
		t.Fail()
	}

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/pv/increment?namespace=hello&key=world", nil)
	router.ServeHTTP(w, req)

	if w.Code != 400 || !strings.Contains(w.Body.String(), "invalid namespace") {
		t.Fail()
	}

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/pv/increment?namespace=test&key=test", nil)
	router.ServeHTTP(w, req)

	var errMsg ErrorMessage
	err := json.Unmarshal(w.Body.Bytes(), &errMsg)
	fmt.Println(w.Body.String(), errMsg)

	if w.Code != 200 || err != nil {
		t.Fail()
	}
	if errMsg.Code != 0 || len(errMsg.Data) != 1 {
		t.Fail()
	}

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/pv/increment?namespace=test&key=test", nil)
	router.ServeHTTP(w, req)

	err = json.Unmarshal(w.Body.Bytes(), &errMsg)
	fmt.Println(w.Body.String(), errMsg)

	if w.Code != 200 || err != nil {
		t.Fail()
	}
	if errMsg.Code != 0 || len(errMsg.Data) != 1 {
		t.Fail()
	}
}

func TestTeardown(t *testing.T) {
	G_db.Delete("test")
	G_db.Delete("test-test")
}
