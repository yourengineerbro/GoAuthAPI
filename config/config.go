package config

import "os"

type Config struct {
	Port   string
	JWTKey []byte
}
// we can fetch the values from environment variables. 
func Load() *Config {
	return &Config{
		Port:   getEnv("PORT", "8080"),
		JWTKey: []byte(getEnv("JWT_KEY", "hjdsGteSiSwt")),
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}