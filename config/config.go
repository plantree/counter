package config

type RedisOptions struct {
}

const (
	SERVER_PORT = ":8080"

	LOG_FILE_PATH = "log"
	LOG_FILE_NAME = "counter-service.log"

	DEFAULT_REDIS_URL = "redis://:@localhost:6379/0"
)
