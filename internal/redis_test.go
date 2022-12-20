package internal

import (
	"os"
	"testing"

	"github.com/plantree/counter/config"
	"github.com/plantree/counter/middleware"
)

func TestNewRedisClient(t *testing.T) {
	logger := middleware.NewLogger()
	defer os.RemoveAll(config.LOG_FILE_PATH)

	if logger == nil {
		t.Error("NewLogger failed")
	}
	if redisClient := NewRedisClient(config.DEFAULT_REDIS_URL, logger); redisClient == nil {
		t.Fail()
	}

}
