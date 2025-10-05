package api

import (
	"log"
	"sample/internal/app/handler"
	"sample/internal/app/repository"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func StartServer() {
	log.Println("Starting server")

	repo, err := repository.NewRepository()
	if err != nil {
		logrus.Error("ошибка инициализации репозитория")
	}

	handler := handler.NewHandler(repo)

	r := gin.Default()
	r.LoadHTMLGlob("templates/*")
	r.Static("/static", "./resources")

	r.GET("/", handler.GetPatients)
	r.GET("/patient/:id", handler.GetPatient)
	r.GET("/insulin_calculation/:id", handler.GetInsulinCalculation)

	r.Run()
	log.Println("Server down")
}

