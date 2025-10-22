package repository

import (
	"sample/internal/app/minio"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Repository struct {
	db          *gorm.DB
	minioClient *minio.MinioClient
}

func NewRepository() (*Repository, error) {
	return &Repository{}, nil
}

func New(dsn string, minioClient *minio.MinioClient) (*Repository, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{}) // подключаемся к БД
	if err != nil {
		return nil, err
	}

	return &Repository{
		db:          db,
		minioClient: minioClient,
	}, nil
}
