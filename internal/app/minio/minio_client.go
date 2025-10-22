package minio

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinioClient struct {
	client     *minio.Client
	bucketName string
	endpoint   string
}

func NewMinioClient(endpoint, accessKey, secretKey, bucketName string, useSSL bool) (*MinioClient, error) {
	// Инициализация клиента MinIO
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("ошибка создания клиента MinIO: %w", err)
	}

	client := &MinioClient{
		client:     minioClient,
		bucketName: bucketName,
		endpoint:   endpoint,
	}

	// Проверяем существование бакета, если нет - создаем
	ctx := context.Background()
	exists, err := minioClient.BucketExists(ctx, bucketName)
	if err != nil {
		return nil, fmt.Errorf("ошибка проверки бакета: %w", err)
	}

	if !exists {
		err = minioClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
		if err != nil {
			return nil, fmt.Errorf("ошибка создания бакета: %w", err)
		}
		log.Printf("Бакет %s создан", bucketName)
	}

	log.Printf("MinIO клиент инициализирован: %s", endpoint)
	return client, nil
}

// Загрузка фото пациента
func (m *MinioClient) UploadPatientPhoto(patientID uint, imagePath string) (string, error) {
	// Проверяем существование файла
	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		return "", fmt.Errorf("файл не найден: %s", imagePath)
	}

	// Генерируем имя файла в MinIO
	objectName := fmt.Sprintf("%d.jpg", patientID)

	// Открываем файл
	file, err := os.Open(imagePath)
	if err != nil {
		return "", fmt.Errorf("ошибка открытия файла: %w", err)
	}
	defer file.Close()

	// Получаем информацию о файле
	fileInfo, err := file.Stat()
	if err != nil {
		return "", fmt.Errorf("ошибка получения информации о файле: %w", err)
	}

	// Загружаем в MinIO
	ctx := context.Background()
	_, err = m.client.PutObject(ctx, m.bucketName, objectName, file, fileInfo.Size(), minio.PutObjectOptions{
		ContentType: "image/jpeg", // Меняйте в зависимости от типа изображения
	})
	if err != nil {
		return "", fmt.Errorf("ошибка загрузки в MinIO: %w", err)
	}

	log.Printf("Фото пациента %d загружено: %s", patientID, objectName)
	return objectName, nil
}

// Получение URL для доступа к фото пациента
func (m *MinioClient) GetPatientPhotoURL(patientID uint, objectName string) string {
	if objectName == "" {
		return ""
	}

	// Для MinIO с публичным доступом
	return fmt.Sprintf("http://%s/%s/%s", m.endpoint, m.bucketName, objectName)
}

// Удаление фото пациента
func (m *MinioClient) DeletePatientPhoto(objectName string) error {
	if objectName == "" {
		return nil
	}

	ctx := context.Background()
	err := m.client.RemoveObject(ctx, m.bucketName, objectName, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("ошибка удаления фото: %w", err)
	}

	log.Printf("Фото удалено: %s", objectName)
	return nil
}

// Загрузка фото из локальной папки для всех пациентов
func (m *MinioClient) UploadAllPatientPhotosFromFolder(localFolderPath string) error {
	files, err := os.ReadDir(localFolderPath)
	if err != nil {
		return fmt.Errorf("ошибка чтения папки: %w", err)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		filename := file.Name()
		// Предполагаем, что имя файла = Patient_ID + расширение
		ext := filepath.Ext(filename)
		nameWithoutExt := strings.TrimSuffix(filename, ext)

		var patientID uint
		_, err := fmt.Sscanf(nameWithoutExt, "%d", &patientID)
		if err != nil {
			log.Printf("Пропускаем файл с некорректным именем: %s", filename)
			continue
		}

		fullPath := filepath.Join(localFolderPath, filename)
		_, err = m.UploadPatientPhoto(patientID, fullPath)
		if err != nil {
			log.Printf("Ошибка загрузки фото для пациента %d: %v", patientID, err)
			continue
		}

		log.Printf("Успешно загружено фото для пациента %d", patientID)
	}

	return nil
}

// Проверка существования фото пациента
func (m *MinioClient) PatientPhotoExists(patientID uint) (bool, error) {
	objectName := fmt.Sprintf("patients/%d.jpg", patientID) // предполагаем jpg

	ctx := context.Background()
	_, err := m.client.StatObject(ctx, m.bucketName, objectName, minio.StatObjectOptions{})
	if err != nil {
		if minio.ToErrorResponse(err).Code == "NoSuchKey" {
			return false, nil
		}
		return false, fmt.Errorf("ошибка проверки фото: %w", err)
	}

	return true, nil
}

// Получение списка всех фото пациентов
func (m *MinioClient) ListAllPatientPhotos() ([]string, error) {
	ctx := context.Background()
	var photos []string

	for object := range m.client.ListObjects(ctx, m.bucketName, minio.ListObjectsOptions{
		Prefix:    "patients/",
		Recursive: true,
	}) {
		if object.Err != nil {
			return nil, fmt.Errorf("ошибка получения списка: %w", object.Err)
		}
		photos = append(photos, object.Key)
	}

	return photos, nil
}
