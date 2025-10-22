package main

import (
	"context"
	"fmt"
	"log"

	_ "sample/docs"
	"sample/internal/app/config"
	"sample/internal/app/dsn"
	"sample/internal/app/handler"
	"sample/internal/app/minio"
	"sample/internal/app/redis"
	"sample/internal/app/repository"
	"sample/internal/pkg"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// @title Insulin Calculation API
// @version 1.0
// @description API для расчета болюсного инсулина
// @contact.name API Support
// @contact.email support@insulin.ru
// @license.name AS IS (NO WARRANTY)
// @host localhost:8080
// @schemes https http
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

func main() {
	router := gin.Default()
	conf, err := config.NewConfig()
	if err != nil {
		logrus.Fatalf("error loading config: %v", err)
	}

	postgresString := dsn.FromEnv()
	fmt.Println(postgresString)

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

	repo, errRep := repository.New(postgresString, minioClient)
	if errRep != nil {
		logrus.Fatalf("error initializing repository: %v", errRep)
	}

	redisClient, err := redis.New(context.Background(), conf.Redis)
	if err != nil {
		logrus.Fatalf("error initializing redis: %v", err)
	}

	hand := handler.NewHandler(repo, redisClient, &conf.JWT)
	application := pkg.NewApp(conf, router, hand)
	application.RunApp()
}
