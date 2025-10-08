package dsn

import (
	"fmt"
	"os"
)

//func FromEnv() string {
//	host := os.Getenv("DB_HOST")
//	if host == "" {
//		return ""
//	}
//	port := os.Getenv("DB_PORT")
//	user := os.Getenv("DB_USER")
//	pass := os.Getenv("DB_PASS")
//	dbname := os.Getenv("DB_NAME")

// И вот мы возвращаем dsn, который необходим для подключения к БД
//	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port, user, pass, dbname)
//}

func FromEnv() string {
	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "5433")
	user := getEnv("DB_USER", "postgres")
	password := getEnv("DB_PASSWORD", "1202")
	dbname := getEnv("DB_NAME", "medic")
	sslmode := getEnv("DB_SSLMODE", "disable")

	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		host, user, password, dbname, port, sslmode)
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

