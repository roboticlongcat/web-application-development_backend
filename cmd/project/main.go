package main

import (
	"fmt"
	"log"

	"sample/internal/app/config"
	"sample/internal/app/dsn"
	"sample/internal/app/handler"
	"sample/internal/app/minio"
	"sample/internal/app/repository"
	"sample/internal/pkg"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func main() {
	router := gin.Default()
	conf, err := config.NewConfig()
	if err != nil {
		logrus.Fatalf("error loading config: %v", err)
	}

	postgresString := dsn.FromEnv()
	fmt.Println(postgresString)

	// Инициализация MinIO клиента ПЕРВОЙ
	minioClient, err := minio.NewMinioClient(
		"localhost:9000", // endpoint
		"minio",          // access key
		"minio124",       // secret key
		"test",           // bucket name
		false,            // useSSL
	)
	if err != nil {
		logrus.Fatalf("error initializing MinIO client: %v", err)
	}
	log.Println("MinIO client initialized successfully")

	// Передаем MinIO клиент в репозиторий
	rep, errRep := repository.New(postgresString, minioClient)
	if errRep != nil {
		logrus.Fatalf("error initializing repository: %v", errRep)
	}

	hand := handler.NewHandler(rep)

	application := pkg.NewApp(conf, router, hand)
	application.RunApp()
}
