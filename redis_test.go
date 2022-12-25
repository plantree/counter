package main

import (
	"fmt"
	"testing"
)

func MockNewRedisClient() *DB {
	logger := NewLogger()

	db := NewRedisClient(DEFAULT_REDIS_URL, logger)
	return db
}

func TestNewRedisClient(t *testing.T) {
	if db := MockNewRedisClient(); db.redisClient == nil {
		t.Fail()
	}
	CleanLog()
}

func TestGet(t *testing.T) {
	db := MockNewRedisClient()
	defer CleanLog()
	ret, err := db.Get("hello")
	if ret != nil || err == nil {
		t.Fail()
	}
	fmt.Println(err)
}

func TestSet(t *testing.T) {
	db := MockNewRedisClient()
	defer CleanLog()
	err := db.Set("test", "yes", true)
	if err != nil {
		fmt.Println(err)
		t.Fail()
	}
	ret, err := db.Get("test")
	if ret.value != "yes" || err != nil {
		fmt.Println(err)
		t.Fail()
	}
}

func TestIncr(t *testing.T) {
	db := MockNewRedisClient()
	defer CleanLog()
	err := db.Set("counter", 0, true)
	ret, err := db.Incr("counter")
	if ret.value.(int64) != 1 || err != nil {
		fmt.Println(err)
		t.Fail()
	}

	// failure case
	ret, err = db.Incr("test")
	fmt.Println(ret, err)
	if err == nil {
		t.Fail()
	}
}

func TestBatchGet(t *testing.T) {
	db := MockNewRedisClient()
	defer CleanLog()

	results, err := db.BatchGet("test", "counter")
	if results == nil || len(results) != 2 || err != nil {
		fmt.Println(err)
		t.Fail()
	}
	fmt.Println(results)
}

func TestDel(t *testing.T) {
	db := MockNewRedisClient()
	defer CleanLog()

	cnt, err := db.Delete("how")
	if cnt != 0 || err != nil {
		fmt.Println(err)
		t.Fail()
	}
	cnt, err = db.Delete("counter", "test")
	fmt.Println(cnt, err)
	if cnt != 2 || err != nil {
		fmt.Println(err)
		t.Fail()
	}

	ret, err := db.Get("test")
	fmt.Println(ret, err)
	if ret != nil || err == nil {
		fmt.Println(err)
		t.Fail()
	}
}

func TestGetPrefixMatchKeys(t *testing.T) {
	db := MockNewRedisClient()
	defer CleanLog()

	for i := 0; i < 33; i++ {
		err := db.Set(fmt.Sprintf("key%d", i), "value", true)
		if err != nil {
			t.Fail()
		}
	}
	keys, err := db.GetPrefixMatchKeys("key*")
	if len(keys) != 33 || err != nil {
		fmt.Println(err)
		t.Fail()
	}
	fmt.Println(keys)
}
