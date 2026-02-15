package utils

import (
	"log"
	"os"
	"time"
)

// GetEnv returns the value of the environment variable or default.
func GetEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// GetDurationEnv parses duration from ENV or returns default.
func GetDurationEnv(key, fallback string) time.Duration {
	val := os.Getenv(key)
	if val == "" {
		val = fallback
	}
	d, err := time.ParseDuration(val)
	if err != nil {
		log.Printf("Invalid duration for %s, using fallback %s", key, fallback)
		d, _ = time.ParseDuration(fallback)
	}
	return d
}
