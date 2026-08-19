package config

import "os"

type Config struct{ HTTPAddr, MySQLDSN, RedisAddr, RedisPassword, AdminUser, AdminPassword string }

func Load() Config {
	return Config{env("HTTP_ADDR", ":8080"), env("MYSQL_DSN", "rental:rental@tcp(mysql:3306)/rental?parseTime=true&charset=utf8mb4"), env("REDIS_ADDR", "redis:6379"), os.Getenv("REDIS_PASSWORD"), env("ADMIN_USERNAME", "admin"), env("ADMIN_PASSWORD", "Admin123!")}
}
func env(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
