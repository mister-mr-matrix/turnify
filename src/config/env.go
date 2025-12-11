package config

import (
	"log"
	"os"
	"strconv"
)

func getEnv(variable string, fallback ...string) string {
	value := os.Getenv(variable)
	if value == "" {
		if len(fallback) == 0 {
			log.Fatalf("Env var %s must be set.", variable)
		} else {
			return fallback[0]
		}
	}
	return value
}

func getEnvInt(variable string, fallback ...int) int {
	value := os.Getenv(variable)
	if value == "" {
		if len(fallback) == 0 {
			log.Fatalf("Env var %s must be set.", variable)
		} else {
			return fallback[0]
		}
	}

	valueI, err := strconv.Atoi(value)
	if err != nil {
		log.Fatalf("Failed to convert string to integer for %s", variable)
	}

	return valueI
}
