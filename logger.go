package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/sirupsen/logrus"
)

type LogFormat struct {
	TimestampFormat string
}

func (f *LogFormat) Format(entry *logrus.Entry) ([]byte, error) {
	var b *bytes.Buffer

	if entry.Buffer != nil {
		b = entry.Buffer
	} else {
		b = &bytes.Buffer{}
	}

	b.WriteByte('[')
	b.WriteString(strings.ToUpper(entry.Level.String()))
	b.WriteString("] ")
	b.WriteString(entry.Time.Format(f.TimestampFormat))
	b.WriteString(" ")

	for _, value := range entry.Data {
		fmt.Fprint(b, value)
		b.WriteString(" ")
	}

	if entry.Message != "" {
		b.WriteString(entry.Message)
		b.WriteString(" ")
	}

	b.WriteByte('\n')
	return b.Bytes(), nil
}

func NewLogger() *logrus.Logger {
	logFilePath := LOG_FILE_PATH
	logFileName := LOG_FILE_NAME

	fileName := path.Join(logFilePath, logFileName)

	if _, err := os.Stat(logFilePath); os.IsNotExist(err) {
		err = os.Mkdir(logFilePath, os.ModePerm)
		if err != nil {
			fmt.Println("Create log director failed. err:", err)
		}
	}
	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		_, err = os.Create(fileName)
		if err != nil {
			fmt.Println("Create log file failed. err:", err)
		}
	}

	// file to write
	f, err := os.OpenFile(fileName, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		fmt.Println("Open file failed. err:", err)
	}

	// new logger
	logger := logrus.New()
	logger.Out = io.MultiWriter(f, os.Stdout)
	logger.SetLevel(logrus.DebugLevel)
	logger.SetFormatter(&LogFormat{
		TimestampFormat: "2006-01-02 15:04:05",
	})

	// set rotatelogs
	_, err = rotatelogs.New(
		strings.Split(fileName, ".")[0]+".%Y%m%d.log",

		rotatelogs.WithLinkName(fileName),

		rotatelogs.WithMaxAge(7*24*time.Hour),

		rotatelogs.WithRotationTime(24*time.Hour),
	)
	if err != nil {
		panic(err)
	}

	return logger
}

func CleanLog() error {
	return os.RemoveAll(LOG_FILE_PATH)
}

func LoggerMiddleware(logger *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		c.Next()

		endTime := time.Now()

		latencyTime := endTime.Sub(startTime)

		reqMethod := c.Request.Method

		reqUri := c.Request.RequestURI

		statusCode := c.Writer.Status()

		clientIp := c.ClientIP()

		logger.WithFields(logrus.Fields{
			"status_code":  statusCode,
			"latency_time": latencyTime,
			"client_ip":    clientIp,
			"req_method":   reqMethod,
			"req_uri":      reqUri,
		}).Info()
	}
}
