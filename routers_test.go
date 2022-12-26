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

func TestIsInt(t *testing.T) {
	if !isInt("123") {
		t.Fail()
	}
	if isInt("123a") {
		t.Fail()
	}
}

func TestConstructKey(t *testing.T) {
	if constructKey("hello", "world") != "hello@world" {
		t.Fail()
	}
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

func TestGetPv(t *testing.T) {
	router := MockRouters()
	defer CleanLog()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/pv/status", nil)
	router.ServeHTTP(w, req)

	if w.Code != 400 || !strings.Contains(w.Body.String(), "need namespace") {
		t.Fail()
	}

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/pv/status?namespace=hello&key=world", nil)
	router.ServeHTTP(w, req)

	if w.Code != 400 || !strings.Contains(w.Body.String(), "invalid namespace") {
		t.Fail()
	}

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/pv/status?namespace=test&key=hello", nil)
	router.ServeHTTP(w, req)

	var errMsg ErrorMessage
	err := json.Unmarshal(w.Body.Bytes(), &errMsg)
	fmt.Println(w.Body.String(), errMsg)

	if w.Code != 400 || err != nil {
		t.Fail()
	}
	if errMsg.Code != 4001 {
		t.Fail()
	}

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/pv/status?namespace=test&key=test", nil)
	router.ServeHTTP(w, req)

	err = json.Unmarshal(w.Body.Bytes(), &errMsg)
	fmt.Println(w.Body.String(), errMsg)

	if w.Code != 200 || err != nil {
		t.Fail()
	}
	if errMsg.Code != 0 || len(errMsg.Data) != 1 {
		t.Fail()
	}

	// get all keys under namespace
	err = G_db.Set("test@test1", 1, true)
	if err != nil {
		t.Fail()
	}
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/pv/status?namespace=test", nil)
	router.ServeHTTP(w, req)

	err = json.Unmarshal(w.Body.Bytes(), &errMsg)
	fmt.Println(w.Body.String(), errMsg)

	if w.Code != 200 || err != nil {
		t.Fail()
	}
	if errMsg.Code != 0 || len(errMsg.Data) != 2 {
		t.Fail()
	}
}

func TestResetPv(t *testing.T) {
	router := MockRouters()
	defer CleanLog()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/pv/reset", nil)
	router.ServeHTTP(w, req)

	if w.Code != 400 || !strings.Contains(w.Body.String(), "need namespace") {
		t.Fail()
	}

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/pv/reset?namespace=test&secret=world&key=test&value=10", nil)
	router.ServeHTTP(w, req)

	var errMsg ErrorMessage
	err := json.Unmarshal(w.Body.Bytes(), &errMsg)
	fmt.Println(w.Body.String(), errMsg)

	if w.Code != 400 || err != nil || errMsg.Code != 4002 {
		t.Fail()
	}

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/pv/reset?namespace=test&secret=test&key=test&value=10", nil)
	router.ServeHTTP(w, req)

	_ = json.Unmarshal(w.Body.Bytes(), &errMsg)
	fmt.Println(w.Body.String(), errMsg)

	result, err := G_db.Get("test@test")
	if w.Code != 200 || err != nil || errMsg.Code != 0 || result.value.(string) != "10" {
		t.Fail()
	}
}

func TestDeletePv(t *testing.T) {
	router := MockRouters()
	defer CleanLog()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/pv/delete", nil)
	router.ServeHTTP(w, req)

	if w.Code != 400 || !strings.Contains(w.Body.String(), "need namespace") {
		t.Fail()
	}

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/pv/delete?namespace=test&secret=test&key=test1", nil)
	router.ServeHTTP(w, req)

	var errMsg ErrorMessage
	err := json.Unmarshal(w.Body.Bytes(), &errMsg)
	fmt.Println(w.Body.String(), errMsg)

	if w.Code != 200 || errMsg.Code != 0 {
		t.Fail()
	}

	result, err := G_db.Get("test@test1")
	if result != nil || err == nil {
		fmt.Println(err)
		t.Fail()
	}
}

func TestTeardown(t *testing.T) {
	G_db.Delete("test")
	G_db.Delete("test@test")
	G_db.Delete("test@test1")
}
