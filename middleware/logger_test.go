package middleware

import (
	"bufio"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/plantree/counter/config"
)

func TestNewLogger(t *testing.T) {
	logger := NewLogger()
	if logger == nil {
		t.Error("NewLogger failed")
	}
	logger.Info("hello world")

	if _, err := os.Stat(config.LOG_FILE_PATH); os.IsNotExist(err) {
		t.Error("log file path not exist")
	}
	defer os.RemoveAll(config.LOG_FILE_PATH)

	fileName := path.Join(config.LOG_FILE_PATH, config.LOG_FILE_NAME)
	file, err := os.Open(fileName)
	if err != nil {
		t.Fail()
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		text := scanner.Text()
		if !strings.Contains(text, "hello world") {
			t.Fail()
		}
		break
	}
}

func TestLoggerMiddleware(t *testing.T) {
	logger := NewLogger()
	if logger == nil {
		t.Error("NewLogger failed")
	}
	r := gin.Default()
	r.Use(LoggerMiddleware(logger))

	if _, err := os.Stat(config.LOG_FILE_PATH); os.IsNotExist(err) {
		t.Error("log file path not exist")
	}
	defer os.RemoveAll(config.LOG_FILE_PATH)
}
