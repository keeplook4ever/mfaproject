package config

import "os"

type Config struct {
	ListenAddr string // ":8080"
	DSN        string // "user:pass@tcp(127.0.0.1:3306)/mfa?parseTime=true&charset=utf8mb4&loc=Local"
	Issuer     string // "MyCompany"
}

func Load() Config {
	return Config{
		ListenAddr: getEnv("LISTEN_ADDR", ":8080"),
		DSN:        getEnv("MYSQL_DSN", "redototp:rpotpWD.,1232@tcp(127.0.0.1:3306)/mfa?parseTime=true&charset=utf8mb4&loc=Local"),
		Issuer:     getEnv("TOTP_ISSUER", "MyCompany"),
	}
}

func getEnv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
