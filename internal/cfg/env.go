// Package cfg handles all the configuration for the project
package cfg

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

func LoadEnv() {
	_ = godotenv.Load()
}

func GetEnv(envName string) (string, error) {
	envValue := os.Getenv(envName)
	if envValue == "" {
		return "", fmt.Errorf("[CONFIG] env %s does not exists", envName)
	}

	return envValue, nil
}
