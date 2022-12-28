package main

import (
	"fmt"
	"testing"
)

func TestMain(t *testing.T) {
	r := Init()
	if r == nil {
		fmt.Println("Init failed")
		t.Fail()
	}
}
