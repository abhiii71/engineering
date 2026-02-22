package config

import (
	"os"
	"strconv"
)

var (
	DatabaseURL string
	SecretKey   string
	Issuer      string
	Port        int
)

func init() {
	DatabaseURL = os.Getenv("DATABASE_URL")
	SecretKey = os.Getenv("SECRET_KEY")
	Issuer = os.Getenv("ISSUER")
	if p := os.Getenv("PORT"); p != "" {
		if port, err := strconv.Atoi(p); err == nil {
			Port = port
		}
	}
	if Port == 0 {
		Port = 8080
	}
}
