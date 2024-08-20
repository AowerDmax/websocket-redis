package config

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	GoAppHost       string
	GoAppPort       int
	RedisHost       string
	RedisPort       int
	MeiliSearchHost string
	MeiliSearchPort int
	IntervalTime    int
	DataQueueKeys   []string
	WsHost          string
	WsPort          int
}

func LoadConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file, using default values")
	}

	return &Config{
		GoAppHost:       getEnv("GO_APP_HOST", "0.0.0.0"),
		GoAppPort:       getEnvAsInt("GO_APP_PORT", 8080),
		RedisHost:       getEnv("REDIS_HOST", "127.0.0.1"),
		RedisPort:       getEnvAsInt("REDIS_PORT", 6379),
		MeiliSearchHost: getEnv("MEILISEARCH_HOST", "localhost"),
		MeiliSearchPort: getEnvAsInt("MEILISEARCH_PORT", 7700),
		IntervalTime:    getEnvAsInt("INTERVAL_TIME", 500),
		DataQueueKeys:   getEnvAsStringSlice("DATA_QUEUE_KEYS", []string{"dialog_manager:chatgpt", "dialog_manager:interviewer", "dialog_manager:rookie"}),
		WsHost:          getEnv("WS_HOST", "0.0.0.0"),
		WsPort:          getEnvAsInt("WS_PORT", 8080),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getEnvAsInt(key string, fallback int) int {
	strValue := getEnv(key, "")
	if value, err := strconv.Atoi(strValue); err == nil {
		return value
	}
	return fallback
}

func getEnvAsStringSlice(key string, fallback []string) []string {
	strValue := getEnv(key, "")
	if strValue == "" {
		return fallback
	}

	strSlice := strings.Split(strValue, ",")
	for i := range strSlice {
		strSlice[i] = strings.TrimSpace(strSlice[i])
	}
	return strSlice
}

// func getEnvAsBool(key string, fallback bool) bool {
// 	strValue := getEnv(key, "")
// 	if value, err := strconv.ParseBool(strValue); err == nil {
// 		return value
// 	}
// 	return fallback
// }

// func getEnvAsIntSlice(key string, fallback []int) []int {
// 	strValue := getEnv(key, "")
// 	if strValue == "" {
// 		return fallback
// 	}
// 	strSlice := strings.Split(strValue, ",")
// 	intSlice := make([]int, len(strSlice))
// 	for i, s := range strSlice {
// 		var err error
// 		intSlice[i], err = strconv.Atoi(strings.TrimSpace(s))
// 		if err != nil {
// 			return fallback
// 		}
// 	}
// 	return intSlice
// }
