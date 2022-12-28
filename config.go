package main

import "time"

const (
	SERVER_PORT = ":8080"

	LOG_FILE_PATH = "log"
	LOG_FILE_NAME = "counter-service.log"

	DEFAULT_REDIS_URL = "redis://:@localhost:6379/0"
	REDIS_KEY_TTL     = 3 * 30 * 24 * 60 * 60 * time.Second
	// for test
	// REDIS_KEY_TTL = 5 * 60 * time.Second
)
